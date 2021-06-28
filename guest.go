package shagoslav

import (
	"fmt"
	"time"
)

type GuestService interface {
	CreateGuest(name string, room *MeetingRoom) (*Guest, error)
	FindGuestByToken(token string) (*Guest, error)
	// DeleteGuest(token string) error
}

type Guest struct {
	RememberToken string
	RoomID        int
	Name          string
	CreatedAt     time.Time
	IsAdmin       bool
}

func NewGuestController(gs GuestService, mrs MeetingRoomService) *GuestController {
	return &GuestController{
		gs:  gs,
		mrs: mrs,
	}
}

type GuestController struct {
	gs  GuestService
	mrs MeetingRoomService
}

// POST /room/signup?name=<name>&token=<token>
func (gc *GuestController) NewGuest(name, token string) (*Guest, error) {
	room, isAdmin, err := gc.mrs.ByToken(token) /////////////////////////////////////////// TODO: I'M HERE!!!!!!!!!!!!!
	if err != nil {
		fmt.Printf("Sorry, we can't find a room with token %s\n"+
			"Error message: %v\n", token, err)
		return nil, err
	}
	gc.gs.CreateGuest()
}
