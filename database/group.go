package database

import (
	"database/sql"
	"fmt"
	"log"
	"shagoslav"
	"time"
)

// NewGroupService returns a new instance of GroupService
func NewGroupService(db *sql.DB) *GroupService {
	grs := GroupService{db: db}
	// grs.DestructiveReset() // TODO: Destructive Reset!!!
	grs.AutoMigrate()
	return &grs
}

// Ensure service implements interface
var _ shagoslav.GroupService = &GroupService{}

type GroupService struct {
	db *sql.DB
}

func (grs *GroupService) AutoMigrate() {
	const queryCreate = `CREATE TABLE IF NOT EXISTS Groups
	(
		id SERIAL PRIMARY KEY,
		name TEXT NOT NULL,
		email TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		remember TEXT UNIQUE NOT NULL,
		is_open BOOLEAN DEFAULT TRUE,
		confirmed BOOLEAN DEFAULT FALSE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		deleted_at TIMESTAMP,
		last_login TIMESTAMP
	);
	`
	_, err := grs.db.Exec(queryCreate)
	if err != nil {
		log.Fatalf("database - cannot create table Groups: %v", err)
	}
}

func (grs *GroupService) DestructiveReset() {
	_, err := grs.db.Exec("DROP TABLE IF EXISTS Groups;")
	if err != nil {
		log.Fatalf("database - cannot drop table Groups: %v", err)
	}
	grs.AutoMigrate()
}

// CreateGroup creates a new Group object and stores it in the db
func (grs *GroupService) CreateGroup(name string, email string, passwordHash string, isOpen bool) (*shagoslav.Group, error) {
	var createdAt time.Time
	var id int
	err := grs.db.QueryRow(`INSERT INTO Groups (name, email, password_hash, is_open)
	VALUES ($1, $2, $3, $4)
	RETURNING (id, created_at);`,
		name, email, passwordHash, isOpen).Scan(&id, createdAt)
	if err != nil {
		return nil, fmt.Errorf("database: GroupService cannot insert a new group into db: %v", err)
	}
	return &shagoslav.Group{
		ID:           id,
		Name:         name,
		AdminEmail:   email,
		PasswordHash: passwordHash,
		IsOpen:       isOpen,
		Confirmed:    false,
		CreatedAt:    createdAt,
	}, nil
}

func (grs *GroupService) ByID(id int) (*shagoslav.Group, error) {
	var g shagoslav.Group
	g.ID = id
	err := grs.db.QueryRow(`SELECT name, email, password_hash, remember, is_open, confirmed, created_at, last_login
	FROM Groups WHERE id = $1`, id).Scan(&g.Name, &g.AdminEmail, &g.PasswordHash, &g.AdminRememberToken,
		&g.IsOpen, &g.Confirmed, &g.CreatedAt, &g.LastLogin)
	if err != nil {
		return nil, fmt.Errorf("GuestService cannot find the guest: %v", err)
	}
	return &g, nil
}

func (grs *GroupService) ByEmail(email string) (*shagoslav.Group, error) {
	var g shagoslav.Group
	g.AdminEmail = email
	err := grs.db.QueryRow(`SELECT id, name, password_hash, remember, is_open, confirmed, created_at, last_login
	FROM Groups WHERE email = $1`, email).Scan(&g.ID, &g.Name, &g.PasswordHash, &g.AdminRememberToken,
		&g.IsOpen, &g.Confirmed, &g.CreatedAt, &g.LastLogin)
	if err != nil {
		return nil, fmt.Errorf("GuestService cannot find the guest: %v", err)
	}
	return &g, nil
}

func (grs *GroupService) ByRemember(remember string) (*shagoslav.Group, error) {
	var g shagoslav.Group
	g.AdminRememberToken = remember
	err := grs.db.QueryRow(`SELECT id, name, email, password_hash, is_open, confirmed, created_at, last_login
	FROM Groups WHERE remember = $1`, remember).Scan(&g.ID, &g.Name, &g.AdminEmail, &g.PasswordHash,
		&g.IsOpen, &g.Confirmed, &g.CreatedAt, &g.LastLogin)
	if err != nil {
		return nil, fmt.Errorf("GuestService cannot find the guest: %v", err)
	}
	return &g, nil
}

func (grs *GroupService) UpdateGroup(g *shagoslav.Group) error {
	_, err := grs.db.Exec(`UPDATE TABLE Groups SET
	name=$1, email=$2, password_hash=$3, is_open=$4, last_login=$5, remember=$6, WHERE id=$7;`,
		g.Name, g.AdminEmail, g.PasswordHash, g.IsOpen, g.LastLogin, g.AdminRememberToken, g.ID)
	if err != nil {
		return err
	}
	return nil
}
