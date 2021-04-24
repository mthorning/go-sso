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
	"log"
	"net/http"
	"os"
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
	utils.CheckErr(err)

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
	r.HandleFunc("/login", handleLogin).Method("POST")
}

var handleLogin = mux.HandlerFunc(func(w http.ResponseWriter, h *http.Request) {
	user, err := getUser(email, password)
	utils.CheckErr(err)
	return jwt.Create(user), nil
})
