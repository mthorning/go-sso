package types

type Session struct {
	ID    string
	Name  string
	Admin bool
}

type User struct {
	ID    string
	Name  string
	Admin bool
	Email string
}

type DbUser struct {
	ID       string
	Name     string
	Password []byte
	Email    string
	Admin    bool
}
