package shagoslav

import "time"

type User struct {
	IDToken   string
	RoomID    int
	Name      string
	CreatedAt time.Time
}

type UserService interface {
	CreateUser(name string, room int)
	FindUserByToken(token string)
	DeleteUser(token string)
}
