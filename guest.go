package shagoslav

import (
	"time"
)

type Guest struct {
	RememberToken string
	RoomID        int
	Name          string
	CreatedAt     time.Time
	IsAdmin       bool
}

type GuestService interface {
	CreateGuest(name string, room *MeetingRoom) (*Guest, error)
	FindGuestByToken(token string) (*Guest, error)
	// DeleteGuest(token string) error
}
