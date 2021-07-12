package database

import (
	"database/sql"
	"fmt"
	"log"
	"shagoslav"
	"shagoslav/rand"
	"time"
)

// Ensure service implements interface
var _ shagoslav.MeetingRoomService = &MeetingRoomService{}

// NewMeetingRoomService returns a new instance of MeetingRoomService
func NewMeetingRoomService(db *sql.DB) *MeetingRoomService {
	ms := MeetingRoomService{db: db}
	ms.AutoMigrate()
	return &ms
}

// MeetingRoomService represents a service for managing meeting rooms
type MeetingRoomService struct {
	db *sql.DB
}

func (ms *MeetingRoomService) AutoMigrate() {
	const queryCreate = `CREATE TABLE IF NOT EXISTS Rooms
	(
		id SERIAL PRIMARY KEY,
		group_id INTEGER UNIQUE,
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

// CreateMeetingRoom creates a new meeting room object and stores it in the database
func (ms *MeetingRoomService) CreateMeetingRoom(name string, groupID int) (*shagoslav.MeetingRoom, error) {
	var id int
	var createdAt time.Time
	guestToken := rand.Token()
	adminToken := rand.Token() //TODO: error check
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

// ByToken retrieves a meeting room object from the database using admin or guest token.
// Returns ErrNotFound if room does not exist
// Also returns a variable isAdmin that indicates whether the visitor is an administrator or a regular guest
func (ms *MeetingRoomService) ByToken(token string) (*shagoslav.MeetingRoom, bool, error) {
	var isAdmin bool = true
	room, err := ms.byAdminToken(token)
	if err == ErrNotFound {
		isAdmin = false
		room, err = ms.byGuestToken(token)
		if err != nil {
			return nil, false, err
		}
	} else if err != nil {
		return nil, false, err
	}
	return room, isAdmin, nil
}

// Retrieves a room from db using admin_token field
func (ms *MeetingRoomService) byAdminToken(token string) (*shagoslav.MeetingRoom, error) {
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

// Retrieves a room from db using guest_token field
func (ms *MeetingRoomService) byGuestToken(token string) (*shagoslav.MeetingRoom, error) {
	mr := new(shagoslav.MeetingRoom)
	mr.GuestToken = token
	err := ms.db.QueryRow(`SELECT id, group_id, name, admin_token, created_at
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
