package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/mthorning/go-sso/jwt"
	"github.com/mthorning/go-sso/types"
	"github.com/mthorning/go-sso/utils"
	"io/ioutil"
	"net/http"
)

type DbUser struct {
	Name     string
	Password string
	Email    string
	Admin    bool
}

func getUser(email string, password string) (types.User, error) {
	jsonFile, err := ioutil.ReadFile("users.json")
	utils.CheckErr(err)

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
	r := mux.NewRouter()
	r.HandleFunc("/login", handleLogin).Methods("POST")
	http.ListenAndServe(":8080", r)
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Error reading form", http.StatusBadRequest)
		return
	}

	email := r.PostFormValue("email")
	password := r.PostFormValue("password")
	user, err := getUser(email, password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	jwt := jwt.New(user)

	json, err := json.Marshal(map[string]string{"jwt": jwt})
	if err != nil {
		http.Error(w, "Error marshalling JSON", http.StatusBadRequest)
	}

	w.WriteHeader(http.StatusOK)
	w.Write(json)
}
