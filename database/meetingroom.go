package database

import (
	"database/sql"
	"fmt"
	"log"
	"shagoslav"
	"shagoslav/rand"
	"time"
)

var _ shagoslav.MeetingRoomService = &MeetingRoomService{}

func NewMeetingRoomService(db *sql.DB) *MeetingRoomService {
	ms := MeetingRoomService{db: db}
	ms.AutoMigrate()
	return &ms
}

type MeetingRoomService struct {
	db *sql.DB
}

func (ms *MeetingRoomService) AutoMigrate() {
	const queryCreate = `CREATE TABLE IF NOT EXISTS Rooms
	(
		id SERIAL PRIMARY KEY,
		group_id INTEGER,
		name TEXT NOT NULL,
		guest_token TEXT,
		admin_token TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		start_time TIMESTAMP
		);`
	_, err := ms.db.Exec(queryCreate)
	if err != nil {
		log.Fatalf("database - cannot create table Rooms: %v", err)
	}
}

func (ms *MeetingRoomService) DestructiveReset() {
	_, err := ms.db.Exec("DROP TABLE IF EXISTS Rooms;")
	if err != nil {
		log.Fatalf("database - cannot drop table Rooms: %v", err)
	}
	ms.AutoMigrate()
}

func (ms *MeetingRoomService) CreateMeetingRoom(name string, groupID int) (*shagoslav.MeetingRoom, error) {
	var id int
	var createdAt time.Time
	guestToken, _ := rand.Token()
	adminToken, _ := rand.Token() //TODO: error check
	// TODO: check if grooupID is valid
	err := ms.db.QueryRow(`INSERT INTO Rooms (name, group_id, guest_token, admin_token)
	VALUES ($1, $2, $3, $4)
	RETURNING id, created_at;`, name, groupID, guestToken, adminToken).Scan(&id, &createdAt)
	if err != nil {
		return nil, fmt.Errorf("MeetingService cannot insert a new room into Rooms: %v", err)
	}
	return &shagoslav.MeetingRoom{
		ID:         id,
		GroupID:    groupID,
		Name:       name,
		GuestToken: guestToken,
		AdminToken: adminToken,
		CreatedAt:  createdAt,
	}, nil
}

// I don't know if this function is needed...
//
// func (ms *MeetingRoomService) ByID(id int) (*shagoslav.MeetingRoom, error) {
// 	mr := new(shagoslav.MeetingRoom)
// 	mr.ID = id
// 	err := ms.db.QueryRow(`SELECT group_id, name, guest_token, admin_token, created_at
// 	FROM Rooms
// 	WHERE id = $1;`, id).Scan(&mr.GroupID, &mr.Name, &mr.GuestToken, &mr.AdminToken, &mr.CreatedAt)
// 	if err != nil {
// 		if err == sql.ErrNoRows {
// 			return nil, ErrNotFound
// 		}
// 		return nil, fmt.Errorf("MeetingService cannot find the room id = %d: %v", id, err)
// 	}
// 	return mr, nil
// }

func (ms *MeetingRoomService) ByAdminToken(token string) (*shagoslav.MeetingRoom, error) {
	mr := new(shagoslav.MeetingRoom)
	mr.AdminToken = token
	err := ms.db.QueryRow(`SELECT id, group_id, name, guest_token, created_at
	FROM Rooms
	WHERE admin_token = $1;`, token).Scan(&mr.ID, &mr.GroupID, &mr.Name, &mr.GuestToken, &mr.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("MeetingService cannot find the room: %v", err)
	}
	return mr, nil
}

func (ms *MeetingRoomService) ByGuestToken(token string) (*shagoslav.MeetingRoom, error) {
	mr := new(shagoslav.MeetingRoom)
	mr.GuestToken = token
	err := ms.db.QueryRow(`SELECT id, group_id, name, guest_token, created_at
	FROM Rooms
	WHERE guest_token = $1;`, token).Scan(&mr.ID, &mr.GroupID, &mr.Name, &mr.AdminToken, &mr.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("MeetingService cannot find the room: %v", err)
	}
	return mr, nil
}
