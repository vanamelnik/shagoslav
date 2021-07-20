package main

import (
	"log"
	"net/http"
	"shagoslav"
	"shagoslav/database"
	"shagoslav/middleware"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

var URL string = "localhost:3000"

func main() {
	db := database.NewDB()
	ms := database.NewMeetingService(db)
	gs := database.NewGuestService(db)

	//mc := shagoslav.NewMeetingRoomsController(ms)
	gc := shagoslav.NewGuestController(gs, ms)

	staticC := shagoslav.NewStatic()
	groupC := shagoslav.NewGroupController(ms)
	router := mux.NewRouter()

	// Routes
	router.Handle("/", staticC.Home)

	// Group routes
	router.HandleFunc("/group", groupC.AccountPage).Methods("GET", "POST") // group account page

	// Meeting routes
	router.HandleFunc("/meeting", middleware.RequireMeetingToken(gc.GuestAtMeeting)).Methods("GET")    // meeting page
	router.HandleFunc("/meeting/newguest", middleware.RequireMeetingToken(gc.NewGuest)).Methods("GET") // new guest form
	router.HandleFunc("/meeting/signup", middleware.RequireMeetingToken(gc.Signup)).Methods("GET")     // create new guest

	log.Fatal(http.ListenAndServe(":3000", router))
}
