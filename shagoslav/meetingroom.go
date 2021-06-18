package shagoslav

import "time"

// MeetingRoom represents an instance of meetingroom
type MeetingRoom struct {
	ID         int
	MemberLink string
	AdminLink  string
	CreatedAt  time.Time
	ExpiresAt  time.Time
}

type MeetingRoomService interface {
	CreateMeetingRoom()
	DeleteMeetingRoom()
}
