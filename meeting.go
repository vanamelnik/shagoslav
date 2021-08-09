package shagoslav

import (
	"time"
)

// Meeting represents an instance of a meeting room.
// Meetings are created by administrators of the groups.
type Meeting struct {
	ID int

	// Meeting title
	Title string

	// id of the group whose administrator created this meeting
	GroupID int

	// GuestToken and AdminToken are randomly generated tokens used in references
	// for ordinary guests and meeting admins
	GuestToken string
	AdminToken string

	// Timestamps for meeting creation and meeting's starting
	CreatedAt time.Time
	StartTime time.Time
}

// MeetingService represents a service for managing meetings
type MeetingService interface {
	// Creates a new meeting object and stores it in the database
	CreateMeeting(name string, groupID int) (*Meeting, error)

	// ByToken retrieves a meeting object from the database using admin or guest token.
	// Also returns a variable isAdmin that indicates whether the visitor is an administrator or a regular guest
	ByToken(token string) (m *Meeting, isAdmin bool, err error)
	ByGroupId(groupId int) (*Meeting, error)
	// DeleteMeeting(id int)

	// AdminsLoggedIn() *[]Guest
}

// // NewMeetingsController creates a new MeetingsController
// func NewMeetingsController(mrs MeetingService) *MeetingsController {
// 	return &MeetingsController{
// 		mrs: mrs,
// 	}
// }

// // MeetingsController provides responses to users' requests for managing meetings
// type MeetingsController struct {
// 	mrs MeetingService
// }
