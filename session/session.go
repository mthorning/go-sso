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

func SetSession(w http.ResponseWriter, r *http.Request, userID string) error {
	s, err := store.Get(r, conf.SessionName)
	if err != nil {
		return err
	}

	s.Values["id"] = userID
	err = s.Save(r, w)
	if err != nil {
		return err
	}
	return nil
}

type NoSessionError struct{}

func (e NoSessionError) Error() string {
	return "No session exists for this user"
}
func GetSession(w http.ResponseWriter, r *http.Request) (types.SessionUser, error) {
	s, err := store.Get(r, conf.SessionName)
	if err != nil {
		return types.SessionUser{}, err
	}
	id, ok := s.Values["id"].(string)
	if !ok {
		return types.SessionUser{}, NoSessionError{}
	}
	return types.SessionUser{
		ID: id,
	}, nil
}

func EndSession(w http.ResponseWriter, r *http.Request) error {
	s, err := store.Get(r, conf.SessionName)
	if err != nil {
		return err
	}
	s.Options.MaxAge = -1
	err = s.Save(r, w)
	if err != nil {
		return err
	}
	return nil
}
