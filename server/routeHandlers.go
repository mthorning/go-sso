package server

import (
	"fmt"
	"github.com/mthorning/go-sso/session"
	"github.com/mthorning/go-sso/types"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
)

type RouteConfig = map[string]interface{}

type AuthRoutes struct {
	SessionUser *types.SessionUser
	Config      RouteConfig
}

func (a AuthRoutes) getData(path string) (interface{}, error) {
	for k, v := range a.Config {
		re, err := regexp.Compile(k)
		if err != nil {
			return nil, err
		}
		if re.MatchString(path) {
			switch c := v.(type) {
			case func(s *types.SessionUser) (interface{}, error):
				return c(a.SessionUser)
			case func(p string, s *types.SessionUser) (interface{}, error):
				return c(path, a.SessionUser)
			default:
				return c, nil
			}
		}
	}
	return nil, nil
}

func (a AuthRoutes) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	sessionUser, err := session.GetSession(w, r)
	if err != nil {
		if _, ok := err.(session.NoSessionError); ok {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		HTMLError(w, r, err.Error(), http.StatusInternalServerError)
		return
	}
	a.SessionUser = &sessionUser

	file := filepath.Clean(r.URL.Path)
	if file == "/" {
		file = "/index"
	}
	templateData, err := a.getData(file)
	if err != nil {
		HTMLError(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	ServeStaticPage(w, r, file, templateData)
}

func NoAuthRoutes(w http.ResponseWriter, r *http.Request) {
	file := filepath.Clean(r.URL.Path)
	ServeStaticPage(w, r, file, nil)
}

func getFilePath(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			re, err := regexp.Compile(`.*\/\[slug\]\.html$`)
			if err != nil {
				return "", err
			}
			if re.MatchString(path) {
				return "", fmt.Errorf("page not found")
			}
			re, err = regexp.Compile(`(.*\/).*(\.html)$`)
			if err != nil {
				return "", err
			}
			return getFilePath(re.ReplaceAllString(path, "$1[slug]$2"))
		} else {
			return "", err
		}
	}
	if info.IsDir() {
		fmt.Println("isDir")
		return "", fmt.Errorf("page not found")
	}
	return path, nil
}

func ServeStaticPage(w http.ResponseWriter, r *http.Request, file string, templateData interface{}) {
	lp := filepath.Join("templates", "layout.html")
	up := filepath.Join("templates", "components.html")

	fp, err := getFilePath(filepath.Join("templates", fmt.Sprintf("%s.html", file)))
	if err != nil {
		HTMLError(w, r, err.Error(), http.StatusNotFound)
		return
	}

	tmpl, err := makeTemplate(w, r, lp, up, fp)
	if err != nil {
		HTMLError(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tmpl.ExecuteTemplate(w, "layout", templateData); err != nil {
		HTMLError(w, r, err.Error(), http.StatusInternalServerError)
		return
	}
}
