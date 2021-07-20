package database

import (
	"database/sql"
	"fmt"
	"log"
	"shagoslav"
	"shagoslav/rand"
	"time"
)

// NewGuestService returns a new instance of GuestService
func NewGuestService(db *sql.DB) *GuestService {
	gs := GuestService{db: db}
	// gs.DestructiveReset() // TODO: Destructive Reset!!!
	gs.AutoMigrate()
	return &gs
}

// Ensure service implements interface
var _ shagoslav.GuestService = &GuestService{}

// GuestService represents a service for managing Guest objects in the database
type GuestService struct {
	db *sql.DB
}

func (gs *GuestService) AutoMigrate() {
	const queryCreate = `CREATE TABLE IF NOT EXISTS Guests
	(
		id SERIAL PRIMARY KEY,
		name TEXT NOT NULL,
		remember TEXT UNIQUE NOT NULL,
		meeting_id INTEGER NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		is_admin BOOLEAN DEFAULT FALSE
	);
	`
	_, err := gs.db.Exec(queryCreate)
	if err != nil {
		log.Fatalf("database - cannot create table Guests: %v", err)
	}
}

func (gs *GuestService) DestructiveReset() {
	_, err := gs.db.Exec("DROP TABLE IF EXISTS Guests;")
	if err != nil {
		log.Fatalf("database - cannot drop table Guests: %v", err)
	}
	gs.AutoMigrate()
}

// CreateGuest creates a new Guest object and stores it in the database
func (gs *GuestService) CreateGuest(name string, meeting *shagoslav.Meeting, isAdmin bool) (*shagoslav.Guest, error) {
	var createdAt time.Time
	rememberToken := rand.Token()

	err := gs.db.QueryRow(`INSERT INTO Guests (name, meeting_id, is_admin, remember)
	VALUES ($1, $2, $3, $4)
	RETURNING created_at;`, name, meeting.ID, isAdmin, rememberToken).Scan(&createdAt)
	if err != nil {
		return nil, fmt.Errorf("GuestService cannot insert a new guest into Guests: %v", err)
	}
	guest := shagoslav.Guest{
		Name:          name,
		MeetingID:     meeting.ID,
		RememberToken: rememberToken,
		CreatedAt:     createdAt,
		IsAdmin:       isAdmin,
	}
	return &guest, nil
}

// ByRemember retrieves a Guest object from the database using guest's remember token
func (gs *GuestService) ByRemember(token string) (*shagoslav.Guest, error) {
	guest := new(shagoslav.Guest)
	err := gs.db.QueryRow(`SELECT name, meeting_id, is_admin, created_at FROM Guests
	WHERE remember = $1;`, token).Scan(&guest.Name, &guest.MeetingID, &guest.IsAdmin, &guest.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("GuestService cannot find the guest: %v", err)
	}
	guest.RememberToken = token
	return guest, nil
}

// GuestsLoggedIn retrieves all guests that logged at the meeting with id = meetingID
func (gs *GuestService) GuestsLoggedIn(meetingID int) (*[]shagoslav.Guest, error) {
	var guests []shagoslav.Guest
	rows, err := gs.db.Query("SELECT remember, name, is_admin, created_at FROM Guests WHERE meeting_id = $1;", meetingID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		g := shagoslav.Guest{MeetingID: meetingID}
		if err := rows.Scan(&g.RememberToken, &g.Name, &g.IsAdmin, &g.CreatedAt); err != nil {
			return nil, err
		}
		guests = append(guests, g)
	}
	return &guests, nil
}

// UpdateGuest updates fields 'name' and 'is_admin' in the Guest object retrieved from the database
// by guest's RememberToken field.
func (gs *GuestService) UpdateGuest(guest *shagoslav.Guest) error {
	_, err := gs.ByRemember(guest.RememberToken)
	if err != nil {
		return err
	}
	_, err = gs.db.Exec(`UPDATE Guests SET name = $1, is_admin = $2 WHERE remember = $3`, guest.Name, guest.IsAdmin, guest.RememberToken)
	return err
}
