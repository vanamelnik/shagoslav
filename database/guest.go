package database

import (
	"database/sql"
	"fmt"
	"log"
	"shagoslav"
	"strings"
	"time"
)

func NewGuestService(db *sql.DB) *GuestService {
	gs := GuestService{db: db}
	gs.AutoMigrate()
	return &gs
}

type GuestService struct {
	db *sql.DB
}

var _ shagoslav.GuestService = &GuestService{}

func (gs *GuestService) AutoMigrate() {
	const queryCreate = `CREATE TABLE IF NOT EXISTS Guests
	(
		id SERIAL PRIMARY KEY,
		name TEXT NOT NULL,
		remember TEXT UNIQUE NOT NULL,
		room_id INTEGER NOT NULL,
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

func (gs *GuestService) CreateGuest(name string, room *shagoslav.MeetingRoom) (*shagoslav.Guest, error) {
	var createdAt time.Time
	rememberToken := "запомните нас такими: " + strings.ToUpper(name) // TODO: make random remember token

	err := gs.db.QueryRow(`INSERT INTO Guests (name, room_id, remember)
	VALUES ($1, $2, $3)
	RETURNING created_at;`, name, room.ID, rememberToken).Scan(&createdAt)
	if err != nil {
		return nil, fmt.Errorf("GuestService cannot insert a new guest into Guests: %v", err)
	}
	guest := shagoslav.Guest{
		Name:          name,
		RoomID:        room.ID,
		RememberToken: rememberToken,
		CreatedAt:     createdAt,
	}
	return &guest, nil
}

func (gs *GuestService) FindGuestByToken(token string) (*shagoslav.Guest, error) {
	guest := new(shagoslav.Guest)
	err := gs.db.QueryRow(`SELECT name, room_id, created_at FROM Guests
	WHERE remember = $1;`, token).Scan(&guest.Name, &guest.RoomID, &guest.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("GuestService cannot find the guest: %v", err)
	}
	guest.RememberToken = token
	return guest, nil
}

// func (gs *GuestService) DeleteGuest()
