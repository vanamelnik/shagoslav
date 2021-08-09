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

func NewGroupController(ms MeetingService, grs GroupService) *GroupController {
	return &GroupController{
		grs:          grs,
		ms:           ms,
		NewGroupView: views.NewView("bootstrap", "views/new_group.gohtml"),
		LoginView:    views.NewView("bootstrap", "views/login.gohtml"),
		AccountView:  views.NewView("bootstrap", "views/account.gohtml"),
	}
}

type GroupController struct {
	grs GroupService
	ms  MeetingService

	NewGroupView *views.View
	LoginView    *views.View
	AccountView  *views.View
}

// getPostFormParams parses PostForm and fills the input form fields using Gorilla Schema
func getPostFormParams(r *http.Request, form interface{}) error {
	if err := r.ParseForm(); err != nil {
		return err
	}
	dec := schema.NewDecoder()
	dec.IgnoreUnknownKeys(true)
	if err := dec.Decode(form, r.PostForm); err != nil {
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
		Name     string `schema:"name"`
		Email    string `schema:"email"`
		Password string `schema:"password"`
		IsOpen   bool   `schema:"isOpen"`
	}
	var f signupForm
	if err := getPostFormParams(r, &f); err != nil {
		log.Printf("GroupController:Signup:Parseform: %v\n", err)
		http.Error(w, "Sorrrrrry... "+err.Error(), http.StatusInternalServerError) // TODO: fix this
		return
	}

	pwdHash, err := bcrypt.GenerateFromPassword([]byte(f.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("GroupController:Signup:bcrypt:%v", err)
		http.Error(w, "Sorry, something went wrong. Please reload the page.", http.StatusInternalServerError)
		return
	}
	f.Password = ""

	// So far the service supports no more than one group attached to a unique e-mail
	if _, err := gc.grs.ByEmail(f.Email); err == nil {
		// TODO: Set alert
		log.Printf("E-mail address %s already registered\n", f.Email)
		http.Redirect(w, r, "/group/signup", http.StatusFound)
		return
	}
	g, err := gc.grs.CreateGroup(f.Name, f.Email, string(pwdHash), f.IsOpen)
	if err != nil {
		log.Printf("GroupService:Signup: %v", err)
		http.Error(w, "Something went wrong, cannot create a new group((", http.StatusInternalServerError)
		return
	}
	log.Printf("GroupController: Successfully created a new group id=%v, name=%s, email=%s, password_hash=%s, created_at=%v",
		g.ID, g.Name, g.AdminEmail, g.PasswordHash, g.CreatedAt)
	if err = gc.signIn(g, w); err != nil {
		log.Printf("GroupService:Signup: %v", err)
		http.Error(w, "Something went wrong, cannot log in((", http.StatusInternalServerError) // TODO: Alert
		return
	}
	http.Redirect(w, r, "/group", http.StatusFound)
}

// Login authenticates admin user, creates remember token and stores it in db & in browser cookie.
//
// POST /group/login
func (gc *GroupController) Login(w http.ResponseWriter, r *http.Request) {
	type loginForm struct {
		Email    string `schema:"email"`
		Password string `schema:"password"`
	}
	var f loginForm
	if err := getPostFormParams(r, &f); err != nil {
		log.Printf("GroupController:Login:Parseform: %v", err)
		http.Error(w, "Sorrrrrry... "+err.Error(), http.StatusInternalServerError) // TODO: fix this
		return
	}
	g, err := gc.grs.ByEmail(f.Email)
	if err != nil {
		log.Printf("GroupService:login:%v", err)
		// TODO: Set Alert wrong email & password
		http.Redirect(w, r, "/group/login", http.StatusFound)
		return
	}
	if err = bcrypt.CompareHashAndPassword([]byte(g.PasswordHash), []byte(f.Password)); err != nil {
		// TODO: Set Alert Wrong email & password
		log.Printf("GroupService:login:%v", err)
		http.Redirect(w, r, "/group/login", http.StatusFound)
		return
	}
	err = gc.signIn(g, w)
	if err != nil {
		log.Printf("GroupService:login:UpdateGroup:%v", err)
		http.Redirect(w, r, "/group/login", http.StatusFound)
		return
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
		Name:    adminRememberCookie,
		Value:   g.AdminRememberToken,
		Expires: time.Now().Add(1 * time.Minute), // TODO: remove this!
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
		Title         string `schema:"title"`
		AcceptOptions bool   `schema:"acceptOptions"`
		Start         string `schema:"start"`
		Duration      string `schema:"duration"`
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
	if err := getPostFormParams(r, &f); err != nil {
		log.Printf("NewMeetingInfo:schema:%v", err)
		http.Redirect(w, r, "/group", http.StatusFound)
		return
	}
	meeting, err := gc.ms.CreateMeeting(f.Title, g.ID)
	if err != nil {
		log.Printf("AccountPage: %v", err)
		http.Error(w, "Something went wrong, cannot create new meeting((", http.StatusInternalServerError)
		return
	}
	log.Printf("AccountPage: Successfuly created a meeting %s\n", meeting.Title)
	http.Redirect(w, r, "/group", http.StatusFound)
}
