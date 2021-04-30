package server

import (
	"github.com/gorilla/sessions"
	"github.com/mthorning/go-sso/config"
	"net/http"
)

type Config struct {
	SessionKey string `default:"devsessionkey"`
}

var (
	conf  Config
	store *sessions.FilesystemStore
)

func init() {
	config.SetConfig(&conf)
	store = sessions.NewFilesystemStore("", []byte(conf.SessionKey))
}

func GetSession(w http.ResponseWriter, r *http.Request, userID string) {
	session, _ := store.Get(r, "sso-session")
	session.Values["id"] = userID
	err := session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
