package shagoslav

import (
	"fmt"
	"time"
)

// MeetingRoom represents an instance of meetingroom
type MeetingRoom struct {
	ID         int
	GroupID    int
	Name       string
	GuestToken string
	AdminToken string
	CreatedAt  time.Time
	StartTime  time.Time
}

type MeetingRoomService interface {
	CreateMeetingRoom(name string, groupID int) (*MeetingRoom, error)

	ByToken(token string) (*MeetingRoom, bool, error)

	// DeleteMeetingRoom(id int)

	// MembersLoggedIn() *[]Guest
	// AdminsLoggedIn() *[]Guest
}

// MeetingRoomsController provides responses to users' requests.
type MeetingRoomsController struct {
	mrs MeetingRoomService
}

// GET /newroom?name=<name>&groupid=<groupId
func (mc *MeetingRoomsController) NewMeetingRoom(name string, groupId int) (guestLink, adminLink string, err error) {
	room, err := mc.mrs.CreateMeetingRoom(name, groupId)
	if err != nil {
		return "", "", err
	}
	guestLink = fmt.Sprintf("/room?token=%s", room.GuestToken)
	adminLink = fmt.Sprintf("/room?token=%s", room.AdminToken)
	return
	// TODO: refresh admin page /group?id=groupId view with filled 'links' fields and message that room is created
}

// GET /room?token=<token>
func (mc *MeetingRoomsController) RoomController(token string) {
	// var isAdmin bool

}
