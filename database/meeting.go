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
var _ shagoslav.MeetingService = &MeetingService{}

// NewMeetingService returns a new instance of MeetingService
func NewMeetingService(db *sql.DB) *MeetingService {
	ms := MeetingService{db: db}
	// ms.AutoMigrate()
	ms.DestructiveReset()
	return &ms
}

// MeetingService represents a service for managing meetings
type MeetingService struct {
	db *sql.DB
}

func (ms *MeetingService) AutoMigrate() {
	const queryCreate = `CREATE TABLE IF NOT EXISTS Meetings
	(
		id SERIAL PRIMARY KEY,
		group_id INTEGER,
		title TEXT NOT NULL,
		guest_token TEXT,
		admin_token TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		start_time TIMESTAMP
		);`
	_, err := ms.db.Exec(queryCreate)
	if err != nil {
		log.Fatalf("database - cannot create table Meetings: %v", err)
	}
}

func (ms *MeetingService) DestructiveReset() {
	_, err := ms.db.Exec("DROP TABLE IF EXISTS Meetings;")
	if err != nil {
		log.Fatalf("database - cannot drop table Meetings: %v", err)
	}
	ms.AutoMigrate()
}

// CreateMeetingRoom creates a new meeting object and stores it in the database
func (ms *MeetingService) CreateMeeting(title string, groupID int) (*shagoslav.Meeting, error) {
	var id int
	var createdAt time.Time
	guestToken := rand.Token()
	adminToken := rand.Token() //TODO: error check
	// TODO: check if grooupID is valid
	err := ms.db.QueryRow(`INSERT INTO Meetings (title, group_id, guest_token, admin_token)
	VALUES ($1, $2, $3, $4)
	RETURNING id, created_at;`, title, groupID, guestToken, adminToken).Scan(&id, &createdAt)
	if err != nil {
		return nil, fmt.Errorf("MeetingService cannot insert a new meeting into Meetings: %v", err)
	}
	return &shagoslav.Meeting{
		ID:         id,
		GroupID:    groupID,
		Title:      title,
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
// 	err := ms.db.QueryRow(`SELECT group_id, title, guest_token, admin_token, created_at
// 	FROM Meetings
// 	WHERE id = $1;`, id).Scan(&mr.GroupID, &mr.title, &mr.GuestToken, &mr.AdminToken, &mr.CreatedAt)
// 	if err != nil {
// 		if err == sql.ErrNoRows {
// 			return nil, ErrNotFound
// 		}
// 		return nil, fmt.Errorf("MeetingService cannot find the meeting id = %d: %v", id, err)
// 	}
// 	return mr, nil
// }

// ByToken retrieves a meeting object from the database using admin or guest token.
// Returns ErrNotFound if the meeting does not exist
// Also returns a variable isAdmin that indicates whether the visitor is an administrator or a regular guest
func (ms *MeetingService) ByToken(token string) (*shagoslav.Meeting, bool, error) {
	var isAdmin bool = true
	meeting, err := ms.byAdminToken(token)
	if err == ErrNotFound {
		isAdmin = false
		meeting, err = ms.byGuestToken(token)
		if err != nil {
			return nil, false, err
		}
	} else if err != nil {
		return nil, false, err
	}
	return meeting, isAdmin, nil
}

// Retrieves a meeting from db using admin_token field
func (ms *MeetingService) byAdminToken(token string) (*shagoslav.Meeting, error) {
	m := new(shagoslav.Meeting)
	m.AdminToken = token
	err := ms.db.QueryRow(`SELECT id, group_id, title, guest_token, created_at
	FROM Meetings
	WHERE admin_token = $1;`, token).Scan(&m.ID, &m.GroupID, &m.Title, &m.GuestToken, &m.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("MeetingService cannot find the meeting: %v", err)
	}
	return m, nil
}

// Retrieves a meeting from db using guest_token field
func (ms *MeetingService) byGuestToken(token string) (*shagoslav.Meeting, error) {
	m := new(shagoslav.Meeting)
	m.GuestToken = token
	err := ms.db.QueryRow(`SELECT id, group_id, title, admin_token, created_at
	FROM Meetings
	WHERE guest_token = $1;`, token).Scan(&m.ID, &m.GroupID, &m.Title, &m.AdminToken, &m.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("MeetingService cannot find the meeting: %v", err)
	}
	return m, nil
}

// Retrieves a meeting from db using guest_token field
func (ms *MeetingService) ByGroupId(groupId int) (*shagoslav.Meeting, error) {
	m := new(shagoslav.Meeting)
	m.GroupID = groupId
	err := ms.db.QueryRow(`SELECT id, title, admin_token, guest_token, created_at
	FROM Meetings
	WHERE group_id = $1;`, groupId).Scan(&m.ID, &m.Title, &m.AdminToken, &m.GuestToken, &m.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("MeetingService cannot find the meeting: %v", err)
	}
	return m, nil
}
