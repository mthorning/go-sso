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

func (s *SessionUser) GetUser(u interface{}) error {
	dsnap, err := firestore.Users.Doc(s.ID).Get(context.Background())
	if err != nil {
		return err
	}
	dsnap.DataTo(u)
	return nil
}
