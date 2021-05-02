package server

import (
	"fmt"
	"github.com/mthorning/go-sso/session"
	"github.com/mthorning/go-sso/types"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
)

type RouteConfig = map[string]interface{}

type AuthRoutes struct {
	SessionUser *types.SessionUser
	Config      RouteConfig
}

func (a AuthRoutes) getData(path string) (interface{}, error) {
	switch c := a.Config[path].(type) {
	case func(s *types.SessionUser) (interface{}, error):
		return c(a.SessionUser)
	default:
		return c, nil
	}
}

func (a AuthRoutes) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	sessionUser, err := session.GetSession(w, r)
	if err != nil {
		if _, ok := err.(session.NoSessionError); ok {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		HTMLError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	a.SessionUser = &sessionUser

	file := filepath.Clean(r.URL.Path)
	if file == "/" {
		file = "/index"
	}
	templateData, err := a.getData(file)
	if err != nil {
		HTMLError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ServeStaticPage(w, r, file, templateData)
}

func NoAuthRoutes(w http.ResponseWriter, r *http.Request) {
	file := filepath.Clean(r.URL.Path)
	ServeStaticPage(w, r, file, nil)
}

func ServeStaticPage(w http.ResponseWriter, r *http.Request, file string, templateData interface{}) {
	lp := filepath.Join("templates", "layout.html")
	up := filepath.Join("templates", "user-form.html")

	fp := filepath.Join("templates", fmt.Sprintf("%s.html", file))

	info, err := os.Stat(fp)
	if err != nil {
		if os.IsNotExist(err) {
			HTMLError(w, "Page Not Found", http.StatusNotFound)
			return
		} else {
			HTMLError(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	if info.IsDir() {
		HTMLError(w, "Page Not Found", http.StatusNotFound)
		return
	}

	tmpl, err := template.ParseFiles(lp, up, fp)
	if err != nil {
		HTMLError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tmpl.ExecuteTemplate(w, "layout", templateData); err != nil {
		HTMLError(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
