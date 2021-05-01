package types

type User struct {
	ID    string
	Name  string
	Email string
	Admin bool
}

type DbUser struct {
	ID       string
	Name     string
	Password []byte
	Email    string
	Admin    bool
}
