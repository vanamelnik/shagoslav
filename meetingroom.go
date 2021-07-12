package shagoslav

import (
	"fmt"
	"log"
	"time"
)

// MeetingRoom represents an instance of a meeting room.
// Rooms are created by administrators of the groups.
type MeetingRoom struct {
	ID int

	// room name
	Name string

	// id of the group whose administrator created this meeting
	GroupID int

	// GuestToken and AdminToken are randomly generated tokens used in references
	// for ordinary guests and meeting admins
	GuestToken string
	AdminToken string

	// Timestamps for room creation and meeting's starting
	CreatedAt time.Time
	StartTime time.Time
}

// MeetingRoomService represents a service for managing meeting rooms
type MeetingRoomService interface {
	// Creates a new meeting room object and stores it in the database
	CreateMeetingRoom(name string, groupID int) (*MeetingRoom, error)

	// ByToken retrieves a meeting room object from the database using admin or guest token.
	// Also returns a variable isAdmin that indicates whether the visitor is an administrator or a regular guest
	ByToken(token string) (mr *MeetingRoom, isAdmin bool, err error)

	// DeleteMeetingRoom(id int)

	// AdminsLoggedIn() *[]Guest
}

// NewMeetingRoomsController creates a new MeetingRoomsController
func NewMeetingRoomsController(mrs MeetingRoomService) *MeetingRoomsController {
	return &MeetingRoomsController{
		mrs: mrs,
	}
}

// MeetingRoomsController provides responses to users' requests for managing meetings
type MeetingRoomsController struct {
	mrs MeetingRoomService
}

// NewMeetingRoom handles query GET /newroom?name=<name>&groupid=<groupId>
// It creates a new meeting using MeetingRoomService and generates guest and admin links
func (mc *MeetingRoomsController) NewMeetingRoom(name string, groupId int) (*MeetingRoom, error) {
	room, err := mc.mrs.CreateMeetingRoom(name, groupId)
	if err != nil {
		return nil, err
	}
	guestLink := fmt.Sprintf("/room?token=%s", room.GuestToken)
	adminLink := fmt.Sprintf("/room?token=%s", room.AdminToken)
	log.Printf("Successfuly created a room %s. Guest link: %s Admin link: %s",
		name, guestLink, adminLink)
	return room, nil
	// TODO: refresh admin page /group?id=groupId view with filled 'links' fields and message that room is created
}
