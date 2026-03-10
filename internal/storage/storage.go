package storage

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"pact/database"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	UploadDir    = "./uploads"
	MaxImageSize = 10 << 20  // 10MB per image
	MaxVideoSize = 100 << 20 // 100MB per video
	MaxAudioSize = 50 << 20  // 50MB per audio
)

var allowedImageTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
}

var allowedVideoTypes = map[string]bool{
	"video/mp4":       true,
	"video/webm":      true,
	"video/quicktime": true,
	"video/x-msvideo": true,
}

var allowedAudioTypes = map[string]bool{
	"audio/mpeg": true,
	"audio/wav":  true,
	"audio/ogg":  true,
	"audio/webm": true,
	"audio/mp4":  true,
	"audio/aac":  true,
}

func Init() error {
	return os.MkdirAll(UploadDir, 0755)
}

func randomFilename() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func taskDir(connectionId, assignedTaskId int64) string {
	return filepath.Join(UploadDir, fmt.Sprintf("%d", connectionId), fmt.Sprintf("%d", assignedTaskId))
}

func SaveFiles(files []*multipart.FileHeader, mediaType string, connectionId, assignedTaskId int64) ([]string, error) {
	if len(files) == 0 {
		return nil, nil
	}

	dir := taskDir(connectionId, assignedTaskId)
	subDir := filepath.Join(dir, mediaType+"s")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		return nil, fmt.Errorf("could not create upload directory: %w", err)
	}

	var maxSize int64
	var allowedTypes map[string]bool
	switch mediaType {
	case "image":
		maxSize = MaxImageSize
		allowedTypes = allowedImageTypes
	case "video":
		maxSize = MaxVideoSize
		allowedTypes = allowedVideoTypes
	case "audio":
		maxSize = MaxAudioSize
		allowedTypes = allowedAudioTypes
	default:
		return nil, fmt.Errorf("unknown media type: %s", mediaType)
	}

	var paths []string
	for _, fh := range files {
		if fh.Size > maxSize {
			return nil, fmt.Errorf("file %q exceeds max size of %dMB", fh.Filename, maxSize/(1<<20))
		}

		f, err := fh.Open()
		if err != nil {
			return nil, fmt.Errorf("could not open uploaded file: %w", err)
		}

		buf := make([]byte, 512)
		n, err := f.Read(buf)
		if err != nil && err != io.EOF {
			f.Close()
			return nil, fmt.Errorf("could not read file header: %w", err)
		}
		contentType := http.DetectContentType(buf[:n])
		if !allowedTypes[contentType] {
			f.Close()
			return nil, fmt.Errorf("file %q has disallowed type %s", fh.Filename, contentType)
		}

		if _, err := f.Seek(0, io.SeekStart); err != nil {
			f.Close()
			return nil, fmt.Errorf("could not seek file: %w", err)
		}

		randName, err := randomFilename()
		if err != nil {
			f.Close()
			return nil, fmt.Errorf("could not generate filename: %w", err)
		}

		ext := filepath.Ext(fh.Filename)
		if ext == "" {
			ext = guessExtension(contentType)
		}
		filename := randName + ext
		destPath := filepath.Join(subDir, filename)

		dst, err := os.Create(destPath)
		if err != nil {
			f.Close()
			return nil, fmt.Errorf("could not create destination file: %w", err)
		}

		if _, err := io.Copy(dst, f); err != nil {
			dst.Close()
			f.Close()
			return nil, fmt.Errorf("could not write file: %w", err)
		}

		dst.Close()
		f.Close()

		relativePath := strings.TrimPrefix(destPath, ".")
		paths = append(paths, relativePath)
	}

	return paths, nil
}

func DeleteTaskFiles(connectionId, assignedTaskId int64) error {
	dir := taskDir(connectionId, assignedTaskId)
	return os.RemoveAll(dir)
}

func guessExtension(contentType string) string {
	switch contentType {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	case "video/mp4":
		return ".mp4"
	case "video/webm":
		return ".webm"
	case "video/quicktime":
		return ".mov"
	case "audio/mpeg":
		return ".mp3"
	case "audio/wav":
		return ".wav"
	case "audio/ogg":
		return ".ogg"
	default:
		return ""
	}
}

// ServeUploadedFile serves files with authorization check
// Path format: /uploads/{connectionId}/{taskId}/images|videos|audios/{filename}
func ServeUploadedFile(w http.ResponseWriter, r *http.Request, userId int) {
	// Extract the request path
	path := strings.TrimPrefix(r.URL.Path, "/uploads/")
	
	// Parse connectionId from path
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	
	connectionId, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		http.Error(w, "Invalid connection ID", http.StatusBadRequest)
		return
	}
	
	// Check if user is the manager of this connection
	queries := database.GetQueries()
	ctx := context.Background()
	
	conn, err := queries.GetActiveConnectionDetails(ctx, connectionId)
	if err != nil {
		http.Error(w, "Connection not found", http.StatusNotFound)
		return
	}
	
	// Only allow manager to view submissions
	if int64(userId) != conn.ManagerID {
		http.Error(w, "Unauthorized - only the manager can view submissions", http.StatusForbidden)
		return
	}
	
	// Serve the file
	fullPath := filepath.Join(UploadDir, path)
	http.ServeFile(w, r, fullPath)
}
