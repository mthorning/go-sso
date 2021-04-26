package types

type User struct {
	Name  string
	Email string
	Admin bool
}

type DbUser struct {
	Name     string
	Password []byte
	Email    string
	Admin    bool
}
