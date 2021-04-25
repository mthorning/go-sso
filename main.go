package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/mthorning/go-sso/config"
	"github.com/mthorning/go-sso/jwt"
	"github.com/mthorning/go-sso/types"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
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

func main() {
	var conf Config
	config.SetConfig(&conf)

	r := mux.NewRouter()
	r.HandleFunc("/login", handleLogin).Methods("POST")
	r.HandleFunc("/authn", handleAuthn).Methods("POST")

	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		serveStaticPage(w, r, nil)
	}).Methods("GET")

	fmt.Println("Serving on port", conf.Port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", conf.Port), r); err != nil {
		log.Fatal(err)
	}
}

func JSONError(w http.ResponseWriter, err string, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"message": err})
}

func JSONResponse(w http.ResponseWriter, response []byte) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

func serveStaticPage(w http.ResponseWriter, r *http.Request, templateData interface{}) {
	lp := filepath.Join("templates", "layout.html")
	fp := filepath.Join("templates", filepath.Clean(r.URL.Path))

	info, err := os.Stat(fp)
	if err != nil {
		if os.IsNotExist(err) {
			http.NotFound(w, r)
			return
		} else {
			JSONError(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	if info.IsDir() {
		fp = filepath.Join("templates", "index.html")
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
