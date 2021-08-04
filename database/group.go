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
	// gs.DestructiveReset() // TODO: Destructive Reset!!!
	grs.AutoMigrate()
	return &grs
}

type GroupService struct {
	db *sql.DB
}

func (grs *GroupService) AutoMigrate() {
	const queryCreate = `CREATE TABLE IF NOT EXISTS Groups
	(
		id SERIAL PRIMARY KEY,
		name TEXT NOT NULL,
		admin_remember TEXT UNIQUE NOT NULL,
		admin_password_hash TEXT NOT NULL,
		is_open BOOLEAN DEFAULT TRUE
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
	err := grs.db.QueryRow(`INSERT INTO Groups (name, email, admin_password_hash, is_open)
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
