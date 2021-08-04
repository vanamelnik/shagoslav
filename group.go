package shagoslav

import (
	"fmt"
	"log"
	"net/http"
	"shagoslav/views"
	"time"

	"github.com/gorilla/schema"
	"golang.org/x/crypto/bcrypt"
)

// TODO: this is a temporary items! GroupService is not implemented yet.
var (
	groupId       = 1
	GroupName     = "Анонимные трубочисты"
	meetingActive = ""
)

type Group struct {
	ID           int
	Name         string
	AdminEmail   string
	PasswordHash string
	IsOpen       bool
	Confirmed    bool

	AdminRememberToken string
	CreatedAt          time.Time
	LastLogin          time.Time
	DeletedAt          time.Time
}

type GroupService interface {
	CreateGroup(name string, email string, passwordHash string, isOpen bool) (*Group, error)
	ByEmail(email string) (*Group, error)
	AdminLogin(email string, password string)
}

func NewGroupController(ms MeetingService) *GroupController {
	return &GroupController{
		AccountView: views.NewView("bootstrap", "views/account.gohtml"),
		ms:          ms,
	}
}

type GroupController struct {
	grs         GroupService
	ms          MeetingService
	AccountView *views.View
}

// getPostFormParams parses PostForm and fills the input form fields using Gorilla Schema
func getPostFromParams(r *http.Request, form interface{}) error {
	if err := r.ParseForm(); err != nil {
		return err
	}
	dec := schema.NewDecoder()
	dec.IgnoreUnknownKeys(true)
	if err := dec.Decode(&form, r.PostForm); err != nil {
		return err
	}
	return nil
}

// Signup hashes admin's password and creates a new 'Group' object in the
// database via GroupService using the data from signup form.
//
// POST /group/signup
func (gc *GroupController) Signup(w http.ResponseWriter, r *http.Request) {
	type signupForm struct {
		name     string `schema:"name"`
		email    string `schema:"email"`
		password string `schema:"password"`
		isOpen   bool   `schema:"isOpen"`
	}
	var f signupForm
	if err := getPostFromParams(r, &f); err != nil {
		log.Printf("GroupController:Signup:Parseform: %v", err)
		http.Error(w, "Sorrrrrry... "+err.Error(), http.StatusInternalServerError) // TODO: fix this
		return
	}

	pwdHash, err := bcrypt.GenerateFromPassword([]byte(f.password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("GroupController:Signup:bcrypt:%v", err)
		http.Error(w, "Sorry, something went wrong. Please reload the page.", http.StatusInternalServerError)
	}
	g, err := gc.grs.CreateGroup(f.name, f.email, string(pwdHash), f.isOpen)
	if err != nil {
		log.Printf("GroupService:Signup: %v", err)
		http.Error(w, "Something went wrong, cannot create a new group((", http.StatusInternalServerError)
	}
	log.Printf("GroupController: Successfully created a new group id=%v, name=%s, email=%s, password_hash=%s, created_at=%v",
		g.ID, g.Name, g.AdminEmail, g.PasswordHash, g.CreatedAt)
	// TODO: login and redirect!!!
}

// Login authenticates admin user, creates remember token and stores it in db & in browser cookie.
//
// POST /group/login
func (gc *GroupController) Login(w http.ResponseWriter, r *http.Request) {
	type loginForm struct {
		email    string `schema:"email"`
		password string `schema:"password"`
	}
	var f loginForm
	if err := getPostFromParams(r, &f); err != nil {
		log.Printf("GroupController:Login:Parseform: %v", err)
		http.Error(w, "Sorrrrrry... "+err.Error(), http.StatusInternalServerError) // TODO: fix this
		return
	}

	g, err := gc.grs.ByEmail(f.email)
	if err != nil {
		log.Printf("GroupService:login:%v", err)
		// TODO: make an alert Not Found
		http.Redirect(w, r, "/group/login", http.StatusFound)
		return
	}
	// Stopped here))
}

// GET /group
func (gc *GroupController) AccountPage(w http.ResponseWriter, r *http.Request) {
	// TODO: check if there is an already created meeting
	log.Printf("account: %v is trying to render account page; meetingActive=%v", r.RemoteAddr, meetingActive)

	if meetingActive == "" {
		gc.AccountView.Render(w, r, views.ViewData{GroupName: GroupName})
	} else {
		m, _, err := gc.ms.ByToken(meetingActive)
		if err != nil {
			log.Printf("AccountPage: look for active meeting: %v", err)
			http.Error(w, "Sorrry... "+err.Error(), http.StatusNotFound)
			return
		}
		gc.AccountView.Render(w, r, views.ViewData{
			GroupName:    GroupName,
			MeetingTitle: m.Title,
			Data: struct {
				GuestLink string
				AdminLink string
			}{
				GuestLink: fmt.Sprintf("/meeting?token=%s", m.GuestToken),
				AdminLink: fmt.Sprintf("/meeting?token=%s", m.AdminToken),
			},
		})
	}
}

// POST /group
func (gc *GroupController) NewMeeting(w http.ResponseWriter, r *http.Request) {
	type meetingForm struct {
		title         string `schema:"title"`
		acceptOptions bool   `schema:"acceptOptions"`
		start         string `schema:"start"`
		duration      string `schema:"duration"`
	}

	var f meetingForm
	// TODO: implement decoding fields "start" and "duration"
	// and update MeetingsService and meetings table in DB
	if err := r.ParseForm(); err != nil {
		log.Printf("NewMeeting:parseform:%v", err)
		gc.AccountPage(w, r)
		return
	}
	dec := schema.NewDecoder()
	dec.IgnoreUnknownKeys(true)
	if err := dec.Decode(&f, r.PostForm); err != nil {
		log.Printf("NewMeetingInfo:schema:%v", err)
		gc.AccountView.Render(w, r, nil)
	}
	meeting, err := gc.ms.CreateMeeting(f.title, groupId)
	if err != nil {
		log.Printf("AccountPage: %v", err)
		http.Error(w, "Something went wrong, cannot create new meeting((", http.StatusInternalServerError)
	}
	log.Printf("AccountPage: Successfuly created a meeting %s\n", f.title)
	meetingActive = meeting.AdminToken
	http.Redirect(w, r, "/group", http.StatusFound)
}
