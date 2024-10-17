package auth

// import (
// 	"context"
// 	"crypto/rand"
// 	"database/sql"
// 	"encoding/base64"
// 	"fmt"
// 	"log"
// 	"net/http"
// 	"pact/internal/database"
// 	"pact/internal/db"
// 	"time"
// )
//
// var sessions = map[string]string{}
//
// func GenSessionId() (string, error) {
// 	byteSlice := make([]byte, 32)
// 	_, err := rand.Read(byteSlice)
// 	if err != nil {
// 		return "", fmt.Errorf("error reading byte slice: %w", err)
// 	}
// 	return base64.URLEncoding.EncodeToString(byteSlice), nil
// }
//
// func SetSession(email string, w http.ResponseWriter) error {
// 	fmt.Println("begun setting session...")
//
// 	queries := database.GetQueries()
//
// 	sessionToken, err := GenSessionId()
// 	if err != nil {
// 		return fmt.Errorf("error generating sessionId")
// 	}
//
// 	user, err := queries.GetUserByEmail(email)
// 	if err != nil {
// 		return fmt.Errorf("get user failed in setSession()")
// 	}
//
// 	session.UserId = int(user.UserID)
// 	session.SessionToken = sessionToken
// 	session.Created, session.Expires = time.Now(), time.Now().Add(time.Hour*24)
//
// 	err = addSession(db, session)
// 	if err != nil {
// 		return fmt.Errorf("error adding session %w", err)
// 	}
//
// 	http.SetCookie(w, &http.Cookie{
// 		Name:    "session_token",
// 		Value:   sessionToken,
// 		Expires: time.Now().Add(24 * time.Hour),
// 	})
// 	fmt.Println("successfully set session")
//
// 	return nil
// }
//
// func addSession(db *sql.DB, session Session) error {
// 	fmt.Printf("UserId: %v\nToken: %v\nCreated: %v\nExpires: %v\n", session.UserId, session.SessionToken, session.Created, session.Expires)
//
// 	query := `INSERT INTO sessions(user_id, session_token, created_at, expires_at) VALUES($1, $2, $3, $4)`
//
// 	_, err := db.Exec(query, session.UserId, session.SessionToken, session.Created, session.Expires)
// 	if err != nil {
// 		return fmt.Errorf("error excecuting query: %w", err)
// 	}
//
// 	fmt.Println(err)
//
// 	return err
// }
//
// func ValidateSession(r *http.Request) (string, error) {
// 	var session Session
// 	db := db.GetDB()
//
// 	cookie, err := r.Cookie("session_token")
// 	if err != nil {
// 		return "", err
// 	}
// 	query := `SELECT * FROM sessions WHERE session_token = ($1)`
// 	row := db.QueryRow(query, cookie.Value)
//
// 	err = row.Scan(&session.Id, &session.UserId, &session.SessionToken, &session.Created, &session.Expires)
// 	if err != nil {
// 		log.Fatal("ValidateSession() failed to discover row in row.Scan()")
// 	}
//
// 	userName, exists := sessions[cookie.Value]
// 	if !exists {
// 		return "", http.ErrNoCookie
// 	}
// 	return userName, nil
// }
//
