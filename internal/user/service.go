package user

import "time"

type User struct {
	UserId      int
	Username    string
	Email       string
	Password    string
	Role        string
	Created_at  time.Time
	Updated_at  time.Time
	Connections map[int]*User // userId as the key, *User as the value
}

// makeUser creates a blank pointer to an uninitialized user.
func makeUser() *User {
	return &User{}
}
