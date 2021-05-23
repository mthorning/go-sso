package types

import (
	"context"
	"github.com/mthorning/go-sso/firestore"
	"time"
)

type User struct {
	ID      string
	Name    string
	Admin   bool
	Email   string
	Created time.Time
}

type DBUser struct {
	ID       string
	Name     string
	Password []byte
	Email    string
	Admin    bool
	Created  time.Time
}

type SessionUser struct {
	ID    string
	Admin bool
}

func (s *SessionUser) GetUser(u interface{}) error {
	dsnap, err := firestore.Users.Doc(s.ID).Get(context.Background())
	if err != nil {
		return err
	}
	dsnap.DataTo(u)
	return nil
}
