package shagoslav

import "time"

type Group struct {
	ID           int
	GroupName    string
	AdminEmail   string
	Password     string
	PasswordHash string
	IsOpen       bool

	AdminRememberToken string
	CreatedAt          time.Time
	LastLogin          time.Time

	SqlGroupService
}

type SqlGroupService interface {
	CreateGroup(name string, email string, password string, isOpen bool) (*Group, error)
	FindGroupByName(name string) (*Group, error)
	AdminLogin(email string, password string)
}
