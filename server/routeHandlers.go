package server

import (
	"fmt"
	"github.com/mthorning/go-sso/session"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
)

type AuthRoutes struct {
	config map[string]func() map[string]interface{}
}

func (a AuthRoutes) getData(path string) map[string]interface{} {
	getDataFunc, ok := a.config[path]
	if !ok {
		return map[string]interface{}{}
	}
	return getDataFunc()
}

func (a AuthRoutes) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s, err := session.GetSession(w, r)
	if err != nil {
		if _, ok := err.(session.NoSessionError); ok {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		HTMLError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	file := filepath.Clean(r.URL.Path)
	if file == "/" {
		file = "index"
	}
	templateData := a.getData(file)
	templateData["session"] = s

	ServeStaticPage(w, r, file, templateData)
}

func NoAuthRoutes(w http.ResponseWriter, r *http.Request) {
	file := filepath.Clean(r.URL.Path)
	ServeStaticPage(w, r, file, nil)
}

func ServeStaticPage(w http.ResponseWriter, r *http.Request, file string, templateData interface{}) {
	lp := filepath.Join("templates", "layout.html")

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

	tmpl, err := template.ParseFiles(lp, fp)
	if err != nil {
		HTMLError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tmpl.ExecuteTemplate(w, "layout", templateData); err != nil {
		HTMLError(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
