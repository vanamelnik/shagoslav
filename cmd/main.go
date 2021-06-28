package main

import (
	"fmt"
	"log"
	"shagoslav"
	"shagoslav/database"

	_ "github.com/lib/pq"
)

var URL string = "localhost:3000"

func main() {
	db := database.NewDB()
	ms := database.NewMeetingRoomService(db)
	names := []string{
		"Alcoholic anonymous",
		"Shopoholic anonymous",
		"VDA",
	}
	rooms := make([]shagoslav.MeetingRoom, 0)
	for _, name := range names {
		room, err := ms.CreateMeetingRoom(name, 112)
		if err != nil {
			log.Println(err)
		}
		rooms = append(rooms, *room)
	}
	fmt.Println(rooms)
}
