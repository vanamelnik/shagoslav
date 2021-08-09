package shagoslav

import (
	"fmt"
	"log"
	"net/http"
	"shagoslav/rand"
	"shagoslav/views"
	"time"

	"github.com/gorilla/schema"
	"golang.org/x/crypto/bcrypt"
)

const adminRememberCookie = "shagoslav_admin"

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
	UpdateGroup(g *Group) error

	ByID(id int) (*Group, error)
	ByEmail(email string) (*Group, error)
	ByRemember(remember string) (*Group, error)
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
func getPostFormParams(r *http.Request, form interface{}) error {
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
	if err := getPostFormParams(r, &f); err != nil {
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
	if err = gc.signIn(g, w); err != nil {
		log.Printf("GroupService:Signup: %v", err)
		http.Error(w, "Something went wrong, cannot log in((", http.StatusInternalServerError) // TODO: Alert
	}
	http.Redirect(w, r, "/group", http.StatusFound)
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
	if err := getPostFormParams(r, &f); err != nil {
		log.Printf("GroupController:Login:Parseform: %v", err)
		http.Error(w, "Sorrrrrry... "+err.Error(), http.StatusInternalServerError) // TODO: fix this
		return
	}
	g, err := gc.grs.ByEmail(f.email)
	if err != nil {
		log.Printf("GroupService:login:%v", err)
		// TODO: Set Alert wrong email & password
		http.Redirect(w, r, "/group/login", http.StatusFound)
		return
	}
	if err = bcrypt.CompareHashAndPassword([]byte(g.PasswordHash), []byte(f.password)); err != nil {
		// TODO: Set Alert Wrong email & password
		log.Printf("GroupService:login:%v", err)
		http.Redirect(w, r, "/group/login", http.StatusFound)
	}
	err = gc.signIn(g, w)
	if err != nil {
		log.Printf("GroupService:login:UpdateGroup%v", err)
		http.Redirect(w, r, "/group/login", http.StatusFound)
	}
	http.Redirect(w, r, "/group", http.StatusFound)
}

// signIn updates LastLogin and AdminRemember fields in the database and creates a cookie
// with admin's freshly generated remember token
func (gc *GroupController) signIn(g *Group, w http.ResponseWriter) error {
	g.AdminRememberToken = rand.Token()
	g.LastLogin = time.Now()
	if err := gc.grs.UpdateGroup(g); err != nil {
		return err
	}
	http.SetCookie(w, &http.Cookie{
		Name:  adminRememberCookie,
		Value: g.AdminRememberToken,
	})
	return nil
}

// GET /group
func (gc *GroupController) AccountPage(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(adminRememberCookie)
	if err != nil {
		http.Redirect(w, r, "/group/login", http.StatusFound)
		return
	}
	g, err := gc.grs.ByRemember(cookie.Value)
	if err != nil {
		log.Printf("AccountPage:remember token in the cookie doesn't match: %v", err)
		http.Redirect(w, r, "/group/login", http.StatusFound) // TODO: Set Alert
		return
	}

	m, err := gc.ms.ByGroupId(g.ID)
	if err != nil {
		gc.AccountView.Render(w, r, views.ViewData{GroupName: g.Name})
	} else {
		gc.AccountView.Render(w, r, views.ViewData{
			MeetingActive: true,
			GroupName:     g.Name,
			MeetingTitle:  m.Title,
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
	// TODO: add middleware to check if there's a cookie and retrieve a Group object
	cookie, err := r.Cookie(adminRememberCookie)
	if err != nil {
		http.Redirect(w, r, "/group/login", http.StatusFound)
		return
	}
	g, err := gc.grs.ByRemember(cookie.Value)
	if err != nil {
		log.Printf("AccountPage:remember token in the cookie doesn't match: %v", err)
		http.Redirect(w, r, "/group/login", http.StatusFound) // TODO: Set Alert
		return
	}

	//TODO: Check unreal case if we already have active meeting...

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
	meeting, err := gc.ms.CreateMeeting(f.title, g.ID)
	if err != nil {
		log.Printf("AccountPage: %v", err)
		http.Error(w, "Something went wrong, cannot create new meeting((", http.StatusInternalServerError)
	}
	log.Printf("AccountPage: Successfuly created a meeting %s\n", meeting.Title)
	http.Redirect(w, r, "/group", http.StatusFound)
}
