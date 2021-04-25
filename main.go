package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/mthorning/go-sso/config"
	"github.com/mthorning/go-sso/firestore"
	"github.com/mthorning/go-sso/jwt"
	"github.com/mthorning/go-sso/types"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type Config struct {
	Port int `default:"8080"`
}

type DbUser struct {
	Name     string
	Password string
	Email    string
	Admin    bool
}

func getUser(email string, password string) (types.User, error) {
	jsonFile, err := ioutil.ReadFile("users.json")
	if err != nil {
		return types.User{}, errors.New("Cannot read users.json")
	}

	var users []DbUser
	err = json.Unmarshal([]byte(jsonFile), &users)
	if err != nil {
		return types.User{}, errors.New("Cannot unmarshall users")
	}

	for _, user := range users {
		// TODO: password hashing & DB obviously
		if user.Email == email && user.Password == password {
			return types.User{
				Email: user.Email,
				Name:  user.Name,
				Admin: user.Admin,
			}, nil
		}
	}

	return types.User{}, errors.New("Username or password does not match")
}

type CatchAll struct{}

func (c CatchAll) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	serveStaticPage(w, r, nil)
}

func main() {
	var conf Config
	config.SetConfig(&conf)

	r := mux.NewRouter()
	r.HandleFunc("/login", handleLogin).Methods("POST")
	r.HandleFunc("/register", handleRegister).Methods("POST")
	r.HandleFunc("/authn", handleAuthn).Methods("POST")

	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	var c CatchAll
	r.PathPrefix("/").Handler(c)

	srv := &http.Server{
		Handler:      r,
		Addr:         fmt.Sprintf("127.0.0.1:%d", conf.Port),
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  30 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}

func JSONError(w http.ResponseWriter, err string, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"message": err})
}

func HTMLError(w http.ResponseWriter, errStr string, code int) {
	lp := filepath.Join("templates", "layout.html")
	ep := filepath.Join("templates", "error.html")

	tmpl, err := template.ParseFiles(lp, ep)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tmpl.ExecuteTemplate(w, "layout",
		struct {
			Code  int
			Error string
		}{code, errStr}); err != nil {
		JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func JSONResponse(w http.ResponseWriter, response []byte) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

func serveStaticPage(w http.ResponseWriter, r *http.Request, templateData interface{}) {
	lp := filepath.Join("templates", "layout.html")
	file := filepath.Clean(r.URL.Path)
	if file == "/" {
		file = "index"
	}

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
		JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tmpl.ExecuteTemplate(w, "layout", templateData); err != nil {
		JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		JSONError(w, "Error reading form", http.StatusBadRequest)
		return
	}

	email := r.PostFormValue("email")
	password := r.PostFormValue("password")
	user, err := getUser(email, password)
	if err != nil {
		r.URL.Path = "/"
		serveStaticPage(w, r, "Email or password incorrect")
		return
	}

	token, err := jwt.New(user)
	if err != nil {
		JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json, err := json.Marshal(map[string]string{"jwt": token})
	if err != nil {
		JSONError(w, "Error marshalling JSON", http.StatusInternalServerError)
		return
	}
	JSONResponse(w, json)
}

func handleRegister(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		JSONError(w, "Error reading form", http.StatusBadRequest)
		return
	}

	email := r.PostFormValue("email")
	password := r.PostFormValue("password")
	name := r.PostFormValue("name")

	_, _, err := firestore.Users.Add(context.Background(), struct {
		Email    string
		Password string
		Name     string
	}{email, password, name})
	if err != nil {
		HTMLError(w, err.Error(), http.StatusBadRequest)
	}
}

func handleAuthn(w http.ResponseWriter, r *http.Request) {
	var rBody map[string]string
	if err := json.NewDecoder(r.Body).Decode(&rBody); err != nil {
		JSONError(w, "Error reading from request body", http.StatusInternalServerError)
		return
	}
	token := rBody["jwt"]

	user, err := jwt.Authenticate(token)
	if err != nil {
		JSONError(w, err.Error(), http.StatusForbidden)
		return
	}

	json, err := json.Marshal(user)
	if err != nil {
		JSONError(w, "Error marshalling JSON", http.StatusInternalServerError)
		return
	}
	JSONResponse(w, json)
}
