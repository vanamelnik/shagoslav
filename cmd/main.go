package main

import (
	"fmt"
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
	fmt.Println("Starting Shagoslav!")
	fmt.Print("Connecting to the database... ")
	db := database.NewDB()
	fmt.Println("Starting services...")
	fmt.Println("\tMeetingService")
	ms := database.NewMeetingService(db)
	fmt.Println("\tGuestService")
	gs := database.NewGuestService(db)
	fmt.Println("\tGroupServiceService")
	grs := database.NewGroupService(db)

	gc := shagoslav.NewGuestController(gs, grs, ms)

	staticC := shagoslav.NewStatic()
	groupC := shagoslav.NewGroupController(ms, grs)

	router := mux.NewRouter()

	// Routes
	router.Handle("/", staticC.Home)

	// Assets
	router.PathPrefix("/assets/").Handler(
		http.StripPrefix("/assets/", http.FileServer(http.Dir("./assets/"))))

	// Group routes
	router.Handle("/group/signup", groupC.NewGroupView).Methods("GET")
	router.HandleFunc("/group/signup", groupC.Signup).Methods("POST")
	router.Handle("/group/login", groupC.LoginView).Methods("GET")
	router.HandleFunc("/group/login", groupC.Login).Methods("POST")
	router.HandleFunc("/group", groupC.AccountPage).Methods("GET") // group account page
	router.HandleFunc("/group", groupC.NewMeeting).Methods("POST") // create new meeting

	// Meeting routes
	router.HandleFunc("/meeting", middleware.RequireMeetingToken(gc.GuestAtMeeting)).Methods("GET")    // meeting page
	router.HandleFunc("/meeting/newguest", middleware.RequireMeetingToken(gc.NewGuest)).Methods("GET") // new guest form
	router.HandleFunc("/meeting/signup", middleware.RequireMeetingToken(gc.Signup)).Methods("GET")     // create new guest

	log.Println("Listening at :3000")
	log.Fatal(http.ListenAndServe(":3000", router))
}
