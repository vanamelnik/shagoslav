package main

import (
	"shagoslav"
	"shagoslav/database"

	_ "github.com/lib/pq"
)

var URL string = "localhost:3000"

func main() {
	db := database.NewDB()
	ms := database.NewMeetingRoomService(db)
	gs := database.NewGuestService(db)

	// mc := shagoslav.NewMeetingRoomController(ms)
	gc := shagoslav.NewGuestController(gs, ms)

	// guest, err := gc.NewGuest("Anakin Skywalker", "bbHo0V9nEMN6Fak2Cph9Xg==")
	// if err != nil {
	// 	log.Println(err)
	// } else {
	// 	log.Printf("Succesfully created a new guest: %+v", guest)
	// }

	gc.GuestInRoom("uOKD9zLshUuEK2o6xMZ4ig==")

	// names := []string{
	// 	"Alcoholic anonymous",
	// 	"Shopoholic anonymous",
	// 	"VDA",
	// }
	// rooms := make([]shagoslav.MeetingRoom, 0)
	// for id, name := range names {
	// 	room, err := mrs.NewMeetingRoom(name, id)
	// 	if err != nil {
	// 		log.Println(err)
	// 	}
	// 	rooms = append(rooms, *room)
	// }
	// fmt.Println("Records successfully inserted")
	// t := rand.Token()
	// tokens := []string{
	// 	rooms[2].GuestToken,
	// 	t,
	// 	rooms[0].AdminToken,
	// }
	// for _, token := range tokens {
	// 	fmt.Printf("\nConnecting to room with token %v. . . \n", token)
	// 	mrs.RoomController(token)
	// }
}
