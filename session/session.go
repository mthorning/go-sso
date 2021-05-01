package session

import (
	"github.com/gorilla/sessions"
	"github.com/mthorning/go-sso/config"
	"github.com/mthorning/go-sso/types"
	"net/http"
)

type Config struct {
	SessionKey  string `default:"devsessionkey"`
	SessionName string `default:"go-sso"`
}

var (
	conf  Config
	store *sessions.FilesystemStore
)

func init() {
	config.SetConfig(&conf)
	store = sessions.NewFilesystemStore("", []byte(conf.SessionKey))
}

func SetSession(w http.ResponseWriter, r *http.Request, user *types.DbUser) error {
	session, err := store.Get(r, conf.SessionName)
	if err != nil {
		return err
	}

	session.Values["id"] = user.ID
	session.Values["name"] = user.Name
	session.Values["admin"] = user.Admin
	err = session.Save(r, w)
	if err != nil {
		return err
	}
	return nil
}

type NoSessionError struct{}

func (e NoSessionError) Error() string {
	return "No session exists for this user"
}
func GetSession(w http.ResponseWriter, r *http.Request) (*types.User, error) {
	session, err := store.Get(r, conf.SessionName)
	if err != nil {
		return nil, err
	}
	id, ok := session.Values["id"].(string)
	if !ok {
		return nil, NoSessionError{}
	}

	name := session.Values["name"].(string)
	admin := session.Values["admin"].(bool)

	user := types.User{
		ID:    id,
		Name:  name,
		Admin: admin,
	}
	return &user, nil
}
