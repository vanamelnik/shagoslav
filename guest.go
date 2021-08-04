package shagoslav

import (
	"fmt"
	"log"
	"net/http"
	"shagoslav/views"
	"time"
)

const (
	guestRememberCookie = "shagoslav_guest" // Cookie name for guest's remember token
	guestCookieExpires  = time.Minute * 3   // TODO: change this interval to 4 hours or something like that
)

// GuestService represents a service for managing Guest objects in the database
type GuestService interface {
	// CreateGuest creates a new Guest object and stores it in the database
	CreateGuest(name string, meeting *Meeting, isAdmin bool) (*Guest, error)

	// ByRemember retrieves a Guest object from the database using guest's remember token
	ByRemember(token string) (*Guest, error)

	// GuestsLoggedIn retrieves all guests that logged at the meeting with id = meetingID
	GuestsLoggedIn(meetingID int) (*[]Guest, error)

	// UpdateGuest updates fields 'name' and 'is_admin' in the Guest object retrieved from the database
	// by guest's RememberToken field.
	UpdateGuest(guest *Guest) error

	// DeleteGuest(token string) error
}

// Guest represents a logged in guest object. Guests are created via the link /meeting/signup?name=<guestName>&token=xxxx
// where 'token' is the admin or guest token from the 'Meeting' object.
// One user may participate in only one meeting at a time.
type Guest struct {
	// Randomly generated remember token stored in a user's cookie
	RememberToken string

	// MeetingID points to the meeting where the guest logged in
	MeetingID int

	// Guest's name
	Name string

	// Timestamp for guest creation
	CreatedAt time.Time

	// IsAdmin shows whether the guest has administrator rights.
	// NOTE: IsAdmin field may be updated when the guest visits the meeting via link of another type.
	IsAdmin bool

	// NOTE: 'Guests' database has a field 'id', but it is only needed for statistics and
	// is not used in the application
}

func NewGuestController(gs GuestService, mrs MeetingService) *GuestController {
	return &GuestController{
		gs:               gs,
		mrs:              mrs,
		SignupView:       views.NewView("bootstrap", "views/new_guest.gohtml"),
		MeetingGuestView: views.NewView("bootstrap", "views/meeting_guest.gohtml"),
		MeetingAdminView: views.NewView("bootstrap", "views/meeting_admin.gohtml"),
	}
}

// GuestController provides responses to users' requests for managing guests and displaying
// views of meeting pages
//
// TODO: может быть те методы, которые отвечают за отображение встречи, следует поместить в
// MeetingController? Или это будет бессмысленное усложнение структуры?
//
type GuestController struct {
	// Services
	gs  GuestService
	mrs MeetingService

	// Views
	SignupView       *views.View
	MeetingGuestView *views.View
	MeetingAdminView *views.View
}

// NewGuest renders guest sign up form
//
// GET /meeting/newguest?token=<MeetingToken>
//
func (gc *GuestController) NewGuest(w http.ResponseWriter, r *http.Request) {
	type signupData struct {
		Token string // We store the token at hidden <input> field
	}

	// With middleware we made sure the token field was present in the request URL
	meetingToken := r.URL.Query().Get("token")
	meeting, _, err := gc.mrs.ByToken(meetingToken)
	if err != nil {
		log.Printf("NewGuest: wrong token %v: %v\n", meetingToken, err)
		http.Redirect(w, r, "/group", http.StatusFound)
		return
	}
	gc.SignupView.Render(w, r, views.ViewData{
		GroupName:    GroupName,
		MeetingTitle: meeting.Title,
		Data:         signupData{Token: meetingToken},
	})
}

// Signup creates a new 'Guest' obect at the meeting with valid GuestToken or AdminToken using
// GuestService, generates guest's remember token and stores it in a user's cookie.
//
// GET /meeting/signup?name=<name>&token=<token>
//
func (gc *GuestController) Signup(w http.ResponseWriter, r *http.Request) {
	meetingToken := r.URL.Query().Get("token") // Thanks middleware, we're sure token is here!..
	name := r.URL.Query().Get("name")          // But we can't be sure about the name.
	if name == "" {
		log.Println("Signup: zero-length name or no name in query. Redirecting back to /newguest")
		http.Redirect(w, r, fmt.Sprintf("/meeting/newguest?token=%s", meetingToken), http.StatusFound)
		return
	}

	meeting, isAdmin, err := gc.mrs.ByToken(meetingToken)
	if err != nil {
		log.Printf("Signup: meetingToken %s is not valid: %v\n", meetingToken, err)
		http.Redirect(w, r, "/group", http.StatusFound)
		return
	}

	g, err := gc.gs.CreateGuest(name, meeting, isAdmin)
	if err != nil {
		log.Printf("Signup: cannot create new guest: %v\n", err)
		http.Error(w, "Sorry, something went wrong...", http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:    guestRememberCookie,
		Value:   g.RememberToken,
		Expires: time.Now().Add(guestCookieExpires),
	})
	log.Printf("Signup: successfully created a new guest %s\n", g.Name)
	// Guest is created, stored in the cookie, so redirecting to the meeting view...
	http.Redirect(w, r, fmt.Sprintf("/meeting?token=%s", meetingToken), http.StatusFound)
}

// GuestAtMeeting shows the main page of the service - meeting view.
// Depending on the token it will be guest's or admin's page.
// User's remember token is stored in a cookie.
//
// GET /meeting?token=<token>
//
func (gc *GuestController) GuestAtMeeting(w http.ResponseWriter, r *http.Request) {
	type meetingInfo struct {
		Guest  *Guest
		Guests *[]Guest
	}
	meetingToken := r.URL.Query().Get("token") // And again we thank middleware))
	meeting, isAdmin, err := gc.mrs.ByToken(meetingToken)
	if err != nil {
		log.Printf("GuestAtMeeting: can't find a meeting with the token %s\n"+
			"Error message: %v\n", meetingToken, err)
		http.Error(w, "Bad Request! Can't find a meeting!", http.StatusBadRequest)
		return
	}

	cookie, err := r.Cookie(guestRememberCookie)
	if err != nil {
		log.Println("GuestAtMeeting: cookie not found. Redirecting to new guest page")
		http.Redirect(w, r, fmt.Sprintf("/meeting/newguest?token=%s", meetingToken), http.StatusFound)
		return
	}
	guest, err := gc.gs.ByRemember(cookie.Value)
	if err != nil {
		log.Printf("GuestAtMeeting: guest with token %v is not found in DB ", cookie.Value)
		http.Redirect(w, r, "/group", http.StatusFound)
		return
	}

	if guest.MeetingID != meeting.ID {
		// TODO: Render ResetGuestView
		fmt.Fprintf(w, "Dear %s, you're already logged at another meeting! One person may not participate several meetings at one time\n",
			guest.Name)
		return
	}
	// if a non-admin user entered the meeting by admin's link, he becomes an admin and vice versa.
	if guest.IsAdmin != isAdmin {
		guest.IsAdmin = isAdmin
		log.Printf("Updating guest %v isAdmin=%v", guest.Name, isAdmin)
		if err = gc.gs.UpdateGuest(guest); err != nil {
			log.Printf("guest service: error when trying to update guest name='%s': %v", guest.Name, err)
		}
	}
	tmpGuests, err := gc.gs.GuestsLoggedIn(guest.MeetingID)
	if err != nil {
		log.Printf("Meeting: cannot get a list of guests: %v", err)
	}

	guests := make([]Guest, 0, len(*tmpGuests)-1)

	for _, g := range *tmpGuests {
		if g.RememberToken != guest.RememberToken {
			guests = append(guests, g) // make a list of guests except you
		}
	}

	viewData := views.ViewData{
		MeetingActive: true,
		GroupName:     GroupName,
		MeetingTitle:  meeting.Title,
		Data:          meetingInfo{Guest: guest, Guests: &guests},
	}
	// render guest or admin meeting view
	if guest.IsAdmin {
		gc.MeetingAdminView.Render(w, r, viewData)
	} else {
		gc.MeetingGuestView.Render(w, r, viewData)
	}
}
