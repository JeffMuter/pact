package auth

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"pact/internal/db"
	"pact/internal/user"
	"time"
)

var sessions = map[string]string{}

type Session struct {
	Id           int
	UserId       int
	SessionToken string
	Created      time.Time
	Expires      time.Time
}

func GenSessionId() (string, error) {
	byteSlice := make([]byte, 32)
	_, err := rand.Read(byteSlice)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(byteSlice), nil
}

func SetSession(email string, w http.ResponseWriter) error {
	sessionToken, err := GenSessionId()
	if err != nil {
		return err
	}

	db := db.GetDB()

	var session Session
	var user user.User

	user, err = user.GetUserByEmail(email)
	if err != nil {
		fmt.Println("get user failed in setSession()")
	}
	session.UserId = user.Id
	session.SessionToken = sessionToken
	session.Created, session.Expires = time.Now(), time.Now().Add(time.Hour*24)

	err = addSession(db, session)
	if err != nil {
		log.Fatal("addSession() err in SetSession()")
	}

	http.SetCookie(w, &http.Cookie{
		Name:    "session_token",
		Value:   sessionToken,
		Expires: time.Now().Add(24 * time.Hour),
	})
	return nil
}

func addSession(db *sql.DB, session Session) error {
	fmt.Printf("UserId: %v\nToken: %v\nCreated: %v\nExpires: %v\n", session.UserId, session.SessionToken, session.Created, session.Expires)
	query := `INSERT INTO sessions(user_id, session_token, created_at, expires_at) VALUES($1, $2, $3, $4)`
	_, err := db.Exec(query, session.UserId, session.SessionToken, session.Created, session.Expires)
	fmt.Println(err)
	return err
}

func ValidateSession(r *http.Request) (string, error) {
	var session Session
	db := database.DatabaseOpen()
	defer db.Close()

	cookie, err := r.Cookie("session_token")
	if err != nil {
		return "", err
	}
	query := `SELECT * FROM sessions WHERE session_token = ($1)`
	row := db.QueryRow(query, cookie.Value)

	err = row.Scan(&session.Id, &session.UserId, &session.SessionToken, &session.Created, &session.Expires)
	if err != nil {
		log.Fatal("ValidateSession() failed to discover row in row.Scan()")
	}

	userName, exists := sessions[cookie.Value]
	if !exists {
		return "", http.ErrNoCookie
	}
	return userName, nil
}
