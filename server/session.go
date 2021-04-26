package server

import (
	"github.com/gorilla/sessions"
	"github.com/mthorning/go-sso/config"
	"github.com/mthorning/go-sso/types"
	"net/http"
)

type Config struct {
	SessionKey string `default:"devsessionkey"`
}

var store *sessions.CookieStore

func init() {
	var conf Config
	config.SetConfig(&conf)
	store = sessions.NewCookieStore([]byte(conf.SessionKey))
}

func GetSession(w http.ResponseWriter, r *http.Request, user types.DbUser) {
	session, _ := store.Get(r, "go-sso")
	session.Values["foo"] = "bar"
	session.Values[42] = 43

	err := session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
