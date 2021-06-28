package shagoslav

import (
	"fmt"
	"log"
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

func NewMeetingRoomController(mrs MeetingRoomService) *MeetingRoomsController {
	return &MeetingRoomsController{
		mrs: mrs,
	}
}

// MeetingRoomsController provides responses to users' requests.
type MeetingRoomsController struct {
	mrs MeetingRoomService
}

// GET /newroom?name=<name>&groupid=<groupId
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

// GET /room?token=<token>
func (mc *MeetingRoomsController) RoomController(token string) {
	room, isAdmin, err := mc.mrs.ByToken(token)
	if err != nil {
		fmt.Printf("Sorry, we can't find a room with token %s\n"+
			"Error message: %v\n", token, err)
		return
	}
	fmt.Println(`****************************************************************
*                   _-===============-_                        *
*               Welcome to our Meeting room!                   *
*                ---===================---                     *
****************************************************************` + "\n")
	fmt.Println("\tRoom info:\n\t----------")
	fmt.Printf("ID:\t\t%v\nGroupID:\t%v\nName:\t\t%v\nGuestToken:\t%v\nAdminToken:\t%v\nCreatedAt:\t%v\n",
		room.ID, room.GroupID, room.Name, room.GuestToken, room.AdminToken, room.CreatedAt)
	if isAdmin {
		fmt.Println("You have admin rights!")
	}
}
