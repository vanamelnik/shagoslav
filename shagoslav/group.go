package shagoslav

type Group struct {
	ID           int
	Name         string
	Email        string
	Password     string
	PasswordHash string
	IsOpen       bool

	AdminRememberToken string
}

func (g *Group) MembersLoggedIn() *[]User
func (g *Group) AdminsLoggedIn() *[]User

type GroupService interface {
	CreateGroup(name string, email string, password string, isOpen bool) (*Group, error)
	FindGroupByName(name string) (*Group, error)
	AdminLogin(email string, password string)
}
