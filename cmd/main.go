package main

import (
	"fmt"
	"log"
	"shagoslav"
	"shagoslav/database"
	"shagoslav/rand"

	_ "github.com/lib/pq"
)

var URL string = "localhost:3000"

func main() {
	db := database.NewDB()
	ms := database.NewMeetingRoomService(db)
	mrs := shagoslav.NewMeetingRoomController(ms)
	names := []string{
		"Alcoholic anonymous",
		"Shopoholic anonymous",
		"VDA",
	}
	rooms := make([]shagoslav.MeetingRoom, 0)
	for id, name := range names {
		room, err := mrs.NewMeetingRoom(name, id)
		if err != nil {
			log.Println(err)
		}
		rooms = append(rooms, *room)
	}
	fmt.Println("Records successfully inserted")
	t, _ := rand.Token()
	tokens := []string{
		rooms[2].GuestToken,
		t,
		rooms[0].AdminToken,
	}
	for _, token := range tokens {
		fmt.Printf("\nConnecting to room with token %v. . . \n", token)
		mrs.RoomController(token)
	}
}
