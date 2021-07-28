package shagoslav

import (
	"fmt"
	"log"
	"net/http"
	"shagoslav/views"
	"time"

	"github.com/gorilla/schema"
)

var groupId = 1 // TODO: this is a temporary item! GroupService is not implemented yet.

type Group struct {
	ID           int
	GroupName    string
	AdminEmail   string
	Password     string
	PasswordHash string
	IsOpen       bool

	AdminRememberToken string
	CreatedAt          time.Time
	LastLogin          time.Time
}

type GroupService interface {
	CreateGroup(name string, email string, password string, isOpen bool) (*Group, error)
	FindGroupByName(name string) (*Group, error)
	AdminLogin(email string, password string)
}

func NewGroupController(ms MeetingService) *GroupController {
	return &GroupController{
		AccountView: views.NewView("bootstrap", "views/account.gohtml"),
		ms:          ms,
	}
}

type GroupController struct {
	ms          MeetingService
	AccountView *views.View
}

type meetingForm struct {
	Title     string `schema:"title"`
	GuestLink string
	AdminLink string
}

// GET POST /group
func (gc *GroupController) AccountPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		gc.AccountView.Render(w, r, nil)
	} else if r.Method == "POST" {
		f := meetingForm{}
		if err := r.ParseForm(); err != nil {
			log.Printf("NewMeetingInfo:parseform:%v", err)
			gc.AccountView.Render(w, r, nil)
		}
		dec := schema.NewDecoder()
		dec.IgnoreUnknownKeys(true)
		if err := dec.Decode(&f, r.PostForm); err != nil {
			log.Printf("NewMeetingInfo:schema:%v", err)
			gc.AccountView.Render(w, r, nil)
		}

		meeting, err := gc.ms.CreateMeeting(f.Title, groupId)
		if err != nil {
			log.Printf("AccountPage: %v", err)
			http.Error(w, "Something went wrong, cannot create new meeting((", http.StatusInternalServerError)
		}
		f.GuestLink = fmt.Sprintf("/meeting?token=%s", meeting.GuestToken)
		f.AdminLink = fmt.Sprintf("/meeting?token=%s", meeting.AdminToken)
		log.Printf("AccountPage: Successfuly created a meeting %s\n", f.Title)

		gc.AccountView.Render(w, r, f)
	} else {
		log.Printf("AccountPage: wrong method: %v", r.Method)
		http.Error(w, "Sorry, something went wrong...", http.StatusBadRequest)
	}
}
