package types

import (
	"context"
	"github.com/mthorning/go-sso/firestore"
	"time"
)

type User struct {
	Name    string
	Admin   bool
	Email   string
	Created time.Time
}

type DbUser struct {
	Name     string
	Password []byte
	Email    string
	Admin    bool
	Created  time.Time
}

type SessionUser struct {
	ID string
}

func (s *SessionUser) GetUser() (User, error) {
	dsnap, err := firestore.Users.Doc(s.ID).Get(context.Background())
	if err != nil {
		return User{}, err
	}
	var user User
	dsnap.DataTo(&user)

	return user, nil
}
