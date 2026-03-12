package buckets

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"pact/database"
	"pact/internal/storage"
	"strconv"
	"strings"
)

// parseTimerFields parses timer form values and returns timer fields with defaults if all are empty
func parseTimerFields(r *http.Request) (timerDays, timerHours, timerMinutes sql.NullInt64) {
	if td := r.FormValue("timer_days"); td != "" {
		val, err := strconv.ParseInt(td, 10, 64)
		if err == nil && val >= 0 {
			timerDays = sql.NullInt64{Int64: val, Valid: true}
		}
	}
	if th := r.FormValue("timer_hours"); th != "" {
		val, err := strconv.ParseInt(th, 10, 64)
		if err == nil && val >= 0 {
			timerHours = sql.NullInt64{Int64: val, Valid: true}
		}
	}
	if tm := r.FormValue("timer_minutes"); tm != "" {
		val, err := strconv.ParseInt(tm, 10, 64)
		if err == nil && val >= 0 {
			timerMinutes = sql.NullInt64{Int64: val, Valid: true}
		}
	}

	// If all timer fields are null/empty, default to 1 day
	if !timerDays.Valid && !timerHours.Valid && !timerMinutes.Valid {
		timerDays = sql.NullInt64{Int64: 1, Valid: true}
		timerHours = sql.NullInt64{Int64: 0, Valid: true}
		timerMinutes = sql.NullInt64{Int64: 0, Valid: true}
	}

	return timerDays, timerHours, timerMinutes
}

func ServeBucketsContent(w http.ResponseWriter, r *http.Request) {
	data := BuildBucketsData(r)
	renderFraction(w, "buckets", data)
}

func HandleCreateTask(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value("userID").(int)
	if userId < 1 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "could not parse form", http.StatusBadRequest)
		return
	}

	title := r.FormValue("title")
	if title == "" {
		http.Error(w, "title is required", http.StatusBadRequest)
		return
	}

	taskType := r.FormValue("type")
	if taskType == "" {
		taskType = "normal"
	}

	defaultPoints, _ := strconv.ParseInt(r.FormValue("points"), 10, 64)
	if defaultPoints <= 0 {
		defaultPoints = 20
	}

	// Parse timer fields (with defaults if empty)
	timerDays, timerHours, timerMinutes := parseTimerFields(r)

	// Calculate duration from timer fields
	durationMinutes := calculateDurationMinutes(timerDays, timerHours, timerMinutes)

	requiresImage := int64(0)
	numImagesRequired := int64(1)
	if r.FormValue("requires_image") == "on" {
		requiresImage = 1
		if numImg, err := strconv.ParseInt(r.FormValue("num_images_required"), 10, 64); err == nil && numImg > 0 {
			numImagesRequired = numImg
		}
	}
	requiresVideo := int64(0)
	numVideosRequired := int64(1)
	if r.FormValue("requires_video") == "on" {
		requiresVideo = 1
		if numVid, err := strconv.ParseInt(r.FormValue("num_videos_required"), 10, 64); err == nil && numVid > 0 {
			numVideosRequired = numVid
		}
	}
	requiresAudio := int64(0)
	numAudioRequired := int64(1)
	if r.FormValue("requires_audio") == "on" {
		requiresAudio = 1
		if numAud, err := strconv.ParseInt(r.FormValue("num_audio_required"), 10, 64); err == nil && numAud > 0 {
			numAudioRequired = numAud
		}
	}

	var minWordCount sql.NullInt64
	if wc := r.FormValue("min_word_count"); wc != "" {
		val, err := strconv.ParseInt(wc, 10, 64)
		if err == nil && val > 0 {
			minWordCount = sql.NullInt64{Int64: val, Valid: true}
		}
	}

	var pointCost sql.NullInt64
	if pc := r.FormValue("point_cost"); pc != "" {
		val, err := strconv.ParseInt(pc, 10, 64)
		if err == nil && val > 0 {
			pointCost = sql.NullInt64{Int64: val, Valid: true}
		}
	}

	isBookmarked := int64(0)
	if r.FormValue("is_bookmarked") == "on" {
		isBookmarked = 1
	}

	var repeatFrequency sql.NullString
	if rf := r.FormValue("repeat_frequency"); rf != "" && rf != "none" {
		repeatFrequency = sql.NullString{String: rf, Valid: true}
	}

	var punishmentTaskID sql.NullInt64
	if ptid := r.FormValue("punishment_task_id"); ptid != "" {
		val, err := strconv.ParseInt(ptid, 10, 64)
		if err == nil && val > 0 {
			punishmentTaskID = sql.NullInt64{Int64: val, Valid: true}
		}
	}

	description := sql.NullString{}
	if desc := r.FormValue("description"); desc != "" {
		description = sql.NullString{String: desc, Valid: true}
	}

	connectionId, err := getActiveConnectionId(r.Context(), userId)
	if err != nil {
		http.Error(w, "no active connection", http.StatusBadRequest)
		return
	}

	assignNow := r.FormValue("assign_now")

	if assignNow == "on" {
		_, err = createAndAssignTask(r.Context(), int64(userId), connectionId, database.CreateTaskParams{
			Title:                  title,
			Description:            description,
			Type:                   taskType,
			DefaultPoints:          defaultPoints,
			DefaultDurationMinutes: durationMinutes,
			TimerDays:              timerDays,
			TimerHours:             timerHours,
			TimerMinutes:           timerMinutes,
			RequiresImage:          requiresImage,
			NumImagesRequired:      numImagesRequired,
			RequiresVideo:          requiresVideo,
			NumVideosRequired:      numVideosRequired,
			RequiresAudio:          requiresAudio,
			NumAudioRequired:       numAudioRequired,
			MinWordCount:           minWordCount,
			PointCost:              pointCost,
			IsBookmarked:           isBookmarked,
			RepeatFrequency:        repeatFrequency,
			PunishmentTaskID:       punishmentTaskID,
		}, durationMinutes, defaultPoints, timerDays, timerHours, timerMinutes)
	} else {
		_, err = createTaskTemplate(r.Context(), int64(userId), database.CreateTaskParams{
			Title:                  title,
			Description:            description,
			Type:                   taskType,
			DefaultPoints:          defaultPoints,
			DefaultDurationMinutes: durationMinutes,
			TimerDays:              timerDays,
			TimerHours:             timerHours,
			TimerMinutes:           timerMinutes,
			RequiresImage:          requiresImage,
			NumImagesRequired:      numImagesRequired,
			RequiresVideo:          requiresVideo,
			NumVideosRequired:      numVideosRequired,
			RequiresAudio:          requiresAudio,
			NumAudioRequired:       numAudioRequired,
			MinWordCount:           minWordCount,
			PointCost:              pointCost,
			IsBookmarked:           isBookmarked,
			RepeatFrequency:        repeatFrequency,
			PunishmentTaskID:       punishmentTaskID,
		})
	}

	if err != nil {
		log.Printf("error creating task: %v", err)
		http.Error(w, "could not create task", http.StatusInternalServerError)
		return
	}

	ServeBucketsContent(w, r)
}

func HandleAssignSavedTask(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value("userID").(int)
	if userId < 1 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	taskId, err := strconv.ParseInt(r.PathValue("task_id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid task id", http.StatusBadRequest)
		return
	}

	err = r.ParseForm()
	if err != nil {
		http.Error(w, "could not parse form", http.StatusBadRequest)
		return
	}

	connectionId, err := getActiveConnectionId(r.Context(), userId)
	if err != nil {
		http.Error(w, "no active connection", http.StatusBadRequest)
		return
	}

	points, _ := strconv.ParseInt(r.FormValue("points"), 10, 64)
	if points <= 0 {
		points = 20
	}

	// Parse timer fields (with defaults if empty)
	timerDays, timerHours, timerMinutes := parseTimerFields(r)

	// Calculate duration from timer fields
	durationMinutes := calculateDurationMinutes(timerDays, timerHours, timerMinutes)

	_, err = assignFromTemplate(r.Context(), int64(userId), taskId, connectionId, durationMinutes, points, timerDays, timerHours, timerMinutes)
	if err != nil {
		log.Printf("error assigning task %d: %v", taskId, err)
		http.Error(w, "could not assign task", http.StatusInternalServerError)
		return
	}

	ServeBucketsContent(w, r)
}

func HandleSaveSubmission(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value("userID").(int)
	if userId < 1 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	assignedTaskId, err := strconv.ParseInt(r.PathValue("assigned_task_id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid task id", http.StatusBadRequest)
		return
	}

	err = r.ParseForm()
	if err != nil {
		http.Error(w, "could not parse form", http.StatusBadRequest)
		return
	}

	submissionText := r.FormValue("submission_text")

	err = saveSubmission(r.Context(), int64(userId), assignedTaskId, submissionText)
	if err != nil {
		log.Printf("error saving submission for task %d: %v", assignedTaskId, err)
		http.Error(w, "could not save submission", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, `<p class="text-emerald-400 text-sm mt-2">Saved</p>`)
}

func HandleSubmitTask(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value("userID").(int)
	if userId < 1 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	assignedTaskId, err := strconv.ParseInt(r.PathValue("assigned_task_id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid task id", http.StatusBadRequest)
		return
	}

	err = r.ParseMultipartForm(200 << 20) // 200MB max
	if err != nil {
		http.Error(w, "could not parse form", http.StatusBadRequest)
		return
	}

	submissionText := r.FormValue("submission_text")

	queries := database.GetQueries()
	ctx := r.Context()
	task, err := queries.GetAssignedTaskById(ctx, assignedTaskId)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprintf(w, `<script>
				alert('This task has been deleted by your manager.');
				setTimeout(function() { 
					htmx.ajax('GET', '/bucketsContent', {target: '#buckets-content', swap: 'outerHTML'}); 
				}, 100);
			</script>`)
			return
		}
		http.Error(w, "could not find task", http.StatusBadRequest)
		return
	}

	var imagePaths, videoPaths, audioPaths []string

	if task.RequiresImage == 1 {
		imageFiles := r.MultipartForm.File["image_files"]
		if int64(len(imageFiles)) < task.NumImagesRequired {
			http.Error(w, fmt.Sprintf("Need %d image(s), received %d", task.NumImagesRequired, len(imageFiles)), http.StatusBadRequest)
			return
		}
		imagePaths, err = storage.SaveFiles(imageFiles, "image", task.ConnectionID, assignedTaskId)
		if err != nil {
			log.Printf("error saving images for task %d: %v", assignedTaskId, err)
			http.Error(w, "could not save images", http.StatusBadRequest)
			return
		}
	}

	if task.RequiresVideo == 1 {
		videoFiles := r.MultipartForm.File["video_files"]
		if int64(len(videoFiles)) < task.NumVideosRequired {
			http.Error(w, fmt.Sprintf("Need %d video(s), received %d", task.NumVideosRequired, len(videoFiles)), http.StatusBadRequest)
			return
		}
		videoPaths, err = storage.SaveFiles(videoFiles, "video", task.ConnectionID, assignedTaskId)
		if err != nil {
			log.Printf("error saving videos for task %d: %v", assignedTaskId, err)
			http.Error(w, "could not save videos", http.StatusBadRequest)
			return
		}
	}

	if task.RequiresAudio == 1 {
		audioFiles := r.MultipartForm.File["audio_files"]
		if int64(len(audioFiles)) < task.NumAudioRequired {
			http.Error(w, fmt.Sprintf("Need %d audio file(s), received %d", task.NumAudioRequired, len(audioFiles)), http.StatusBadRequest)
			return
		}
		audioPaths, err = storage.SaveFiles(audioFiles, "audio", task.ConnectionID, assignedTaskId)
		if err != nil {
			log.Printf("error saving audio for task %d: %v", assignedTaskId, err)
			http.Error(w, "could not save audio", http.StatusBadRequest)
			return
		}
	}

	imageJSON, _ := json.Marshal(imagePaths)
	videoJSON, _ := json.Marshal(videoPaths)
	audioJSON, _ := json.Marshal(audioPaths)

	err = submitTask(r.Context(), int64(userId), assignedTaskId, submissionText, string(imageJSON), string(videoJSON), string(audioJSON))
	if err != nil {
		log.Printf("error submitting task %d: %v", assignedTaskId, err)
		if strings.Contains(err.Error(), "could not find assigned task") {
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprintf(w, `<script>
				alert('This task has been deleted by your manager. Your submission text has not been saved. Please copy it now if needed.');
				setTimeout(function() { 
					htmx.ajax('GET', '/bucketsContent', {target: '#buckets-content', swap: 'outerHTML'}); 
				}, 100);
			</script>`)
			return
		}
		http.Error(w, "could not submit task", http.StatusBadRequest)
		return
	}

	ServeBucketsContent(w, r)
}

func HandleApproveTask(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value("userID").(int)
	if userId < 1 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	assignedTaskId, err := strconv.ParseInt(r.PathValue("assigned_task_id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid task id", http.StatusBadRequest)
		return
	}

	err = approveTask(r.Context(), int64(userId), assignedTaskId)
	if err != nil {
		log.Printf("error approving task %d: %v", assignedTaskId, err)
		http.Error(w, "could not approve task", http.StatusInternalServerError)
		return
	}

	ServeBucketsContent(w, r)
}

func HandleDisapproveTask(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value("userID").(int)
	if userId < 1 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	assignedTaskId, err := strconv.ParseInt(r.PathValue("assigned_task_id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid task id", http.StatusBadRequest)
		return
	}

	err = disapproveTask(r.Context(), int64(userId), assignedTaskId)
	if err != nil {
		log.Printf("error disapproving task %d: %v", assignedTaskId, err)
		http.Error(w, "could not disapprove task", http.StatusInternalServerError)
		return
	}

	ServeBucketsContent(w, r)
}

func HandlePurchaseReward(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value("userID").(int)
	if userId < 1 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	taskId, err := strconv.ParseInt(r.PathValue("task_id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid task id", http.StatusBadRequest)
		return
	}

	connectionId, err := getActiveConnectionId(r.Context(), userId)
	if err != nil {
		http.Error(w, "no active connection", http.StatusBadRequest)
		return
	}

	_, err = purchaseReward(r.Context(), int64(userId), connectionId, taskId)
	if err != nil {
		log.Printf("error purchasing reward %d: %v", taskId, err)
		http.Error(w, "could not purchase reward", http.StatusBadRequest)
		return
	}

	ServeBucketsContent(w, r)
}

func HandleDeleteTask(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value("userID").(int)
	if userId < 1 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	taskId, err := strconv.ParseInt(r.PathValue("task_id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid task id", http.StatusBadRequest)
		return
	}

	err = deleteTaskTemplate(r.Context(), int64(userId), taskId)
	if err != nil {
		log.Printf("error deleting task %d: %v", taskId, err)
		http.Error(w, "could not delete task", http.StatusInternalServerError)
		return
	}

	ServeBucketsContent(w, r)
}

func HandleCreateReward(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value("userID").(int)
	if userId < 1 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "could not parse form", http.StatusBadRequest)
		return
	}

	title := r.FormValue("title")
	if title == "" {
		http.Error(w, "title is required", http.StatusBadRequest)
		return
	}

	pointCost, _ := strconv.ParseInt(r.FormValue("point_cost"), 10, 64)
	if pointCost <= 0 {
		http.Error(w, "point cost must be positive", http.StatusBadRequest)
		return
	}

	defaultPoints, _ := strconv.ParseInt(r.FormValue("points"), 10, 64)
	durationMinutes, _ := strconv.ParseInt(r.FormValue("duration_minutes"), 10, 64)
	if durationMinutes <= 0 {
		durationMinutes = 1440
	}

	requiresImage := int64(0)
	numImagesRequired := int64(1)
	if r.FormValue("requires_image") == "on" {
		requiresImage = 1
		if numImg, err := strconv.ParseInt(r.FormValue("num_images_required"), 10, 64); err == nil && numImg > 0 {
			numImagesRequired = numImg
		}
	}
	requiresVideo := int64(0)
	numVideosRequired := int64(1)
	if r.FormValue("requires_video") == "on" {
		requiresVideo = 1
		if numVid, err := strconv.ParseInt(r.FormValue("num_videos_required"), 10, 64); err == nil && numVid > 0 {
			numVideosRequired = numVid
		}
	}
	requiresAudio := int64(0)
	numAudioRequired := int64(1)
	if r.FormValue("requires_audio") == "on" {
		requiresAudio = 1
		if numAud, err := strconv.ParseInt(r.FormValue("num_audio_required"), 10, 64); err == nil && numAud > 0 {
			numAudioRequired = numAud
		}
	}

	var minWordCount sql.NullInt64
	if wc := r.FormValue("min_word_count"); wc != "" {
		val, err := strconv.ParseInt(wc, 10, 64)
		if err == nil && val > 0 {
			minWordCount = sql.NullInt64{Int64: val, Valid: true}
		}
	}

	description := sql.NullString{}
	if desc := r.FormValue("description"); desc != "" {
		description = sql.NullString{String: desc, Valid: true}
	}

	_, err = createTaskTemplate(r.Context(), int64(userId), database.CreateTaskParams{
		Title:                  title,
		Description:            description,
		Type:                   "reward",
		DefaultPoints:          defaultPoints,
		DefaultDurationMinutes: durationMinutes,
		RequiresImage:          requiresImage,
		NumImagesRequired:      numImagesRequired,
		RequiresVideo:          requiresVideo,
		NumVideosRequired:      numVideosRequired,
		RequiresAudio:          requiresAudio,
		NumAudioRequired:       numAudioRequired,
		MinWordCount:           minWordCount,
		PointCost:              sql.NullInt64{Int64: pointCost, Valid: true},
		IsBookmarked:           0,
	})
	if err != nil {
		log.Printf("error creating reward: %v", err)
		http.Error(w, "could not create reward", http.StatusInternalServerError)
		return
	}

	ServeBucketsContent(w, r)
}

func HandleUpdateReward(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value("userID").(int)
	if userId < 1 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	taskId, err := strconv.ParseInt(r.PathValue("task_id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid task id", http.StatusBadRequest)
		return
	}

	err = r.ParseForm()
	if err != nil {
		http.Error(w, "could not parse form", http.StatusBadRequest)
		return
	}

	title := r.FormValue("title")
	if title == "" {
		http.Error(w, "title is required", http.StatusBadRequest)
		return
	}

	pointCost, _ := strconv.ParseInt(r.FormValue("point_cost"), 10, 64)
	if pointCost <= 0 {
		http.Error(w, "point cost must be positive", http.StatusBadRequest)
		return
	}

	defaultPoints, _ := strconv.ParseInt(r.FormValue("points"), 10, 64)
	durationMinutes, _ := strconv.ParseInt(r.FormValue("duration_minutes"), 10, 64)
	if durationMinutes <= 0 {
		durationMinutes = 1440
	}

	requiresImage := int64(0)
	numImagesRequired := int64(1)
	if r.FormValue("requires_image") == "on" {
		requiresImage = 1
		if numImg, err := strconv.ParseInt(r.FormValue("num_images_required"), 10, 64); err == nil && numImg > 0 {
			numImagesRequired = numImg
		}
	}
	requiresVideo := int64(0)
	numVideosRequired := int64(1)
	if r.FormValue("requires_video") == "on" {
		requiresVideo = 1
		if numVid, err := strconv.ParseInt(r.FormValue("num_videos_required"), 10, 64); err == nil && numVid > 0 {
			numVideosRequired = numVid
		}
	}
	requiresAudio := int64(0)
	numAudioRequired := int64(1)
	if r.FormValue("requires_audio") == "on" {
		requiresAudio = 1
		if numAud, err := strconv.ParseInt(r.FormValue("num_audio_required"), 10, 64); err == nil && numAud > 0 {
			numAudioRequired = numAud
		}
	}

	var minWordCount sql.NullInt64
	if wc := r.FormValue("min_word_count"); wc != "" {
		val, err := strconv.ParseInt(wc, 10, 64)
		if err == nil && val > 0 {
			minWordCount = sql.NullInt64{Int64: val, Valid: true}
		}
	}

	description := sql.NullString{}
	if desc := r.FormValue("description"); desc != "" {
		description = sql.NullString{String: desc, Valid: true}
	}

	err = updateTaskTemplate(r.Context(), int64(userId), taskId, database.UpdateTaskParams{
		Title:                  title,
		Description:            description,
		Type:                   "reward",
		DefaultPoints:          defaultPoints,
		DefaultDurationMinutes: durationMinutes,
		RequiresImage:          requiresImage,
		NumImagesRequired:      numImagesRequired,
		RequiresVideo:          requiresVideo,
		NumVideosRequired:      numVideosRequired,
		RequiresAudio:          requiresAudio,
		NumAudioRequired:       numAudioRequired,
		MinWordCount:           minWordCount,
		PointCost:              sql.NullInt64{Int64: pointCost, Valid: true},
		IsBookmarked:           0,
		RepeatFrequency:        sql.NullString{},
		PunishmentTaskID:       sql.NullInt64{},
		TaskID:                 taskId,
		ManagerID:              int64(userId),
	})
	if err != nil {
		log.Printf("error updating reward %d: %v", taskId, err)
		http.Error(w, "could not update reward", http.StatusInternalServerError)
		return
	}

	ServeBucketsContent(w, r)
}

func HandleDeleteReward(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value("userID").(int)
	if userId < 1 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	taskId, err := strconv.ParseInt(r.PathValue("task_id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid task id", http.StatusBadRequest)
		return
	}

	err = deleteTaskTemplate(r.Context(), int64(userId), taskId)
	if err != nil {
		log.Printf("error deleting reward %d: %v", taskId, err)
		http.Error(w, "could not delete reward", http.StatusInternalServerError)
		return
	}

	ServeBucketsContent(w, r)
}

func HandleDeleteAssignedTask(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value("userID").(int)
	if userId < 1 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	assignedTaskId, err := strconv.ParseInt(r.PathValue("assigned_task_id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid task id", http.StatusBadRequest)
		return
	}

	err = deleteAssignedTask(r.Context(), int64(userId), assignedTaskId)
	if err != nil {
		log.Printf("error deleting assigned task %d: %v", assignedTaskId, err)
		http.Error(w, "could not delete assigned task", http.StatusInternalServerError)
		return
	}

	ServeBucketsContent(w, r)
}

func HandleEditAssignedTask(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value("userID").(int)
	if userId < 1 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	assignedTaskId, err := strconv.ParseInt(r.PathValue("assigned_task_id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid task id", http.StatusBadRequest)
		return
	}

	err = r.ParseForm()
	if err != nil {
		http.Error(w, "could not parse form", http.StatusBadRequest)
		return
	}

	title := r.FormValue("title")
	if title == "" {
		http.Error(w, "title is required", http.StatusBadRequest)
		return
	}

	description := sql.NullString{}
	if desc := r.FormValue("description"); desc != "" {
		description = sql.NullString{String: desc, Valid: true}
	}

	points, _ := strconv.ParseInt(r.FormValue("points"), 10, 64)
	if points <= 0 {
		points = 20
	}

	// Parse timer fields (with defaults if empty)
	timerDays, timerHours, timerMinutes := parseTimerFields(r)

	// Calculate duration from timer fields
	durationMinutes := calculateDurationMinutes(timerDays, timerHours, timerMinutes)

	requiresImage := int64(0)
	numImagesRequired := int64(1)
	if r.FormValue("requires_image") == "on" {
		requiresImage = 1
		if numImg, err := strconv.ParseInt(r.FormValue("num_images_required"), 10, 64); err == nil && numImg > 0 {
			numImagesRequired = numImg
		}
	}
	requiresVideo := int64(0)
	numVideosRequired := int64(1)
	if r.FormValue("requires_video") == "on" {
		requiresVideo = 1
		if numVid, err := strconv.ParseInt(r.FormValue("num_videos_required"), 10, 64); err == nil && numVid > 0 {
			numVideosRequired = numVid
		}
	}
	requiresAudio := int64(0)
	numAudioRequired := int64(1)
	if r.FormValue("requires_audio") == "on" {
		requiresAudio = 1
		if numAud, err := strconv.ParseInt(r.FormValue("num_audio_required"), 10, 64); err == nil && numAud > 0 {
			numAudioRequired = numAud
		}
	}

	var minWordCount sql.NullInt64
	if wc := r.FormValue("min_word_count"); wc != "" {
		val, err := strconv.ParseInt(wc, 10, 64)
		if err == nil && val > 0 {
			minWordCount = sql.NullInt64{Int64: val, Valid: true}
		}
	}

	err = updateAssignedTask(r.Context(), int64(userId), assignedTaskId, title, description, points, durationMinutes, timerDays, timerHours, timerMinutes, requiresImage, numImagesRequired, requiresVideo, numVideosRequired, requiresAudio, numAudioRequired, minWordCount)
	if err != nil {
		log.Printf("error updating assigned task %d: %v", assignedTaskId, err)
		http.Error(w, "could not update task", http.StatusBadRequest)
		return
	}

	ServeBucketsContent(w, r)
}
