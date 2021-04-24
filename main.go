package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/mthorning/go-sso/jwt"
	"github.com/mthorning/go-sso/types"
	"github.com/mthorning/go-sso/utils"
	"io/ioutil"
	"log"
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
		// TODO: password hashing
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

func newLogin(email string, password string) (string, error) {
	user, err := getUser(email, password)
	utils.CheckErr(err)

	return jwt.Create(user), nil
}

func main() {
	args := os.Args[1:]
	if len(args) != 2 {
		log.Fatal("Need email and password")
	}

	email := args[0]
	password := args[1]

	jwt, err := newLogin(email, password)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(jwt)
}
