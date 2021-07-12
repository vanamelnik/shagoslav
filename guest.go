package shagoslav

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

// now cookies are not implemented, so we use this stub
var cookieToken string = ""

// GuestService represents a service for managing Guest objects in the database
type GuestService interface {
	// CreateGuest creates a new Guest object and stores it in the database
	CreateGuest(name string, room *MeetingRoom, isAdmin bool) (*Guest, error)

	// ByRemember retrieves a Guest object from the database using guest's remember token
	ByRemember(token string) (*Guest, error)

	// GuestsLoggedIn retrieves all guests that logged in the room with id = roomID
	GuestsLoggedIn(roomID int) (*[]Guest, error)
	// UpdateGuest updates fields 'name' and 'is_admin' in the Guest object retrieved from the database
	// by guest's RememberToken field.
	UpdateGuest(guest *Guest) error

	// DeleteGuest(token string) error
}

// Guest represents a logged in guest obect. Guests are created via the link /room/signup?name=<guestName>&token=xxxx
// where 'token' is the admin or guest token from the 'Room' object.
type Guest struct {
	// 'Guests' database has a field 'id', but it is only needed for statistics and is not used in the application

	// Randomly generated remember token stored in a user's cookie
	RememberToken string

	// RoomID points to room where the guest logged in
	RoomID int

	// Guest's name
	Name string

	// Timestamp for guest creation
	CreatedAt time.Time

	// IsAdmin shows whether the guest has administrator rights
	// NOTE: IdAdmin field may be updated when the guest visit the meeting via link of another type.
	IsAdmin bool
}

func NewGuestController(gs GuestService, mrs MeetingRoomService) *GuestController {
	return &GuestController{
		gs:  gs,
		mrs: mrs,
	}
}

// GuestController provides responses to users' requests for managing guests
type GuestController struct {
	gs  GuestService
	mrs MeetingRoomService
}

// NewGuest handles a query POST /room/signup?name=<name>&token=<token>
// It creates a new 'Guest' obect in the room with valid GuestToken or AdminToken using GuestService,
// generates guest's remember token and stores it in a user's cookie.
func (gc *GuestController) NewGuest(name, roomToken string) (*Guest, error) {
	room, isAdmin, err := gc.mrs.ByToken(roomToken)
	if err != nil {
		fmt.Printf("Sorry, there's no room with the token %s\n"+
			"Error message: %v\n", roomToken, err)
		return nil, err
	}

	g, err := gc.gs.CreateGuest(name, room, isAdmin)
	if err != nil {
		return nil, err
	}
	// TODO: We need to save guest's remember token in a cookie at client's side
	// and then redirect to /room?token=<roomToken>
	cookieToken = g.RememberToken
	return g, nil
}

// GuestInRoom handles GET /room?token=<token> query. It shows guest's or admin's view
func (gc *GuestController) GuestInRoom(roomToken string) {
	room, isAdmin, err := gc.mrs.ByToken(roomToken)
	if err != nil {
		fmt.Printf("Sorry, we can't find a room with the token %s\n"+
			"Error message: %v\n", roomToken, err)
		return
	}
	guest, ok := gc.guestFromCookie()
	if !ok {
		// TODO: this is a stub. We should redirect to the NewGuest view
		fmt.Printf("Welcome to %s meeting room! Please sign in!\nWhat's your name: ", room.Name)
		in := bufio.NewReader(os.Stdin)
		name, _ := in.ReadString('\n')
		name = strings.TrimSuffix(name, "\n")
		guest, err = gc.NewGuest(name, roomToken)
		if err != nil {
			log.Printf("Sorry, internal error: %v", err)
		}
		cookieToken = guest.RememberToken
	} else if guest.RoomID != room.ID {
		fmt.Printf("Dear %s, you're already logged in another room! One person may not participate several meetings at one time\n",
			guest.Name)
		return
	}
	if guest.IsAdmin != isAdmin {
		guest.IsAdmin = isAdmin
		if err = gc.gs.UpdateGuest(guest); err != nil {
			log.Printf("guest controller error when trying to update guest name='%s': %v", guest.Name, err)
		}
	}
	fmt.Println(`****************************************************************
*                   _-===============-_                        *
*               Welcome to our Meeting room!                   *
*                ---===================---                     *
****************************************************************` + "\n")
	fmt.Println("\tRoom info:\n\t----------")
	fmt.Printf("ID:\t\t%v\nGroupID:\t%v\nName:\t\t%v\nGuestToken:\t%v\nAdminToken:\t%v\nCreatedAt:\t%v\n",
		room.ID, room.GroupID, room.Name, room.GuestToken, room.AdminToken, room.CreatedAt)
	fmt.Println("\n\tInfo about you:")
	fmt.Printf("Name:\t\t%v\nRoomID:\t\t%v\nRememberToken:\t%v\nCreatedAt:\t%v\nIsAdmin:\t\t%v\n",
		guest.Name, guest.RoomID, guest.RememberToken, guest.CreatedAt, guest.IsAdmin)
	if isAdmin {
		fmt.Println("You have admin rights!")
	}
	guests, err := gc.gs.GuestsLoggedIn(room.ID)
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Printf("\nThere are %d guests in our room:\n", len(*guests))
	for n, g := range *guests {
		fmt.Printf("%d\t%s", n+1, g.Name)
		if g.IsAdmin {
			fmt.Print("\tadmin")
		}
		fmt.Println()
	}

}

// guestFromCookie retrieves a Guest object from the database with remember token from the user's cookie using
// GuestService
func (gc *GuestController) guestFromCookie() (*Guest, bool) {
	g, err := gc.gs.ByRemember(cookieToken)
	if err != nil {
		return nil, false
	}
	return g, true
}
