package main

import (
	"encoding/json"
	"errors"
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
	r.HandleFunc("/authn", handleAuthn).Methods("POST")
	http.ListenAndServe(":8080", r)
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

func handleLogin(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		JSONError(w, "Error reading form", http.StatusBadRequest)
		return
	}

	email := r.PostFormValue("email")
	password := r.PostFormValue("password")
	user, err := getUser(email, password)
	if err != nil {
		JSONError(w, err.Error(), http.StatusForbidden)
		return
	}

	token := jwt.New(user)

	json, err := json.Marshal(map[string]string{"jwt": token})
	if err != nil {
		JSONError(w, "Error marshalling JSON", http.StatusBadRequest)
	}
	JSONResponse(w, json)
}

func handleAuthn(w http.ResponseWriter, r *http.Request) {
	var rBody map[string]string
	if err := json.NewDecoder(r.Body).Decode(&rBody); err != nil {
		JSONError(w, "Error reading from request body", http.StatusBadRequest)
	}
	token := rBody["jwt"]

	user, err := jwt.Authenticate(token)
	if err != nil {
		JSONError(w, err.Error(), http.StatusForbidden)
		return
	}

	json, err := json.Marshal(user)
	if err != nil {
		JSONError(w, "Error marshalling JSON", http.StatusBadRequest)
		return
	}
	JSONResponse(w, json)

}
