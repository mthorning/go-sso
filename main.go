package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/mitchellh/mapstructure"
	"github.com/mthorning/go-sso/config"
	"github.com/mthorning/go-sso/firestore"
	"github.com/mthorning/go-sso/jwt"
	"github.com/mthorning/go-sso/server"
	"github.com/mthorning/go-sso/types"
	"golang.org/x/crypto/bcrypt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type Config struct {
	Port int `default:"8080"`
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
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

func getJWT(w http.ResponseWriter, user types.User) {
	// Not sure about this yet
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

func handleLogin(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		HTMLError(w, "Error reading form", http.StatusBadRequest)
		return
	}

	email := r.PostFormValue("email")
	password := r.PostFormValue("password")

	var invalid = func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = "/"
		serveStaticPage(w, r, "Email or password incorrect")
	}

	dsnap, err := firestore.Users.Doc(email).Get(context.Background())
	if err != nil {
		invalid(w, r)
		return
	}

	var user types.DbUser
	mapstructure.Decode(dsnap.Data(), &user)

	if err = bcrypt.CompareHashAndPassword(user.Password, []byte(password)); err != nil {
		invalid(w, r)
		return
	}

	server.GetSession(w, r, user)
	http.Redirect(w, r, "/welcome", http.StatusFound)
}

func handleRegister(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		JSONError(w, "Error reading form", http.StatusBadRequest)
		return
	}

	email := r.PostFormValue("email")
	password := r.PostFormValue("password")
	name := r.PostFormValue("name")

	// TODO Check for duplicate email

	pw, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		HTMLError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = firestore.Users.Doc(email).Set(context.Background(), struct {
		Email    string
		Password []byte
		Name     string
	}{email, pw, name})
	if err != nil {
		HTMLError(w, err.Error(), http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
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
