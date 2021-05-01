package server

import (
	"context"
	"encoding/json"
	"github.com/mthorning/go-sso/firestore"
	"github.com/mthorning/go-sso/jwt"
	"github.com/mthorning/go-sso/session"
	"github.com/mthorning/go-sso/types"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/api/iterator"
	"net/http"
	"time"
)

func HandleLogin(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		HTMLError(w, "Error reading form", http.StatusBadRequest)
		return
	}

	email := r.PostFormValue("email")
	password := r.PostFormValue("password")

	var sendError = func(errorMessage string) {
		ServeStaticPage(w, r, "error", map[string]string{
			"email": email,
			"error": errorMessage,
		})
	}
	if email == "" {
		sendError("Please enter an email address")
		return
	}
	if password == "" {
		sendError("Please enter a password")
		return
	}

	query := firestore.Users.Where("Email", "==", email)
	iter := query.Documents(context.Background())
	defer iter.Stop()
	doc, err := iter.Next()
	if err == iterator.Done {
		sendError("Email or password incorrect")
		return
	}
	if err != nil {
		HTMLError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var user types.DbUser
	doc.DataTo(&user)

	if err = bcrypt.CompareHashAndPassword(user.Password, []byte(password)); err != nil {
		sendError("Email or password incorrect")
		return
	}

	if err := session.SetSession(w, r, &user); err != nil {
		HTMLError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

func HandleRegister(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		JSONError(w, "Error reading form", http.StatusBadRequest)
		return
	}

	email := r.PostFormValue("email")
	password := r.PostFormValue("password")
	name := r.PostFormValue("name")

	var sendError = func(errorMessage string) {
		ServeStaticPage(w, r, "error", map[string]string{
			"name":  name,
			"email": email,
			"error": errorMessage,
		})
	}
	if name == "" {
		sendError("Please provide a name")
		return
	}
	if email == "" {
		sendError("Please enter an email address")
		return
	}
	if password == "" {
		sendError("Please enter a password")
		return
	}

	query := firestore.Users.Where("Email", "==", email)
	iter := query.Documents(context.Background())
	defer iter.Stop()
	for {
		_, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			HTMLError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		sendError("Email address already taken")
		return
	}

	pw, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		HTMLError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, _, err = firestore.Users.Add(context.Background(), struct {
		Email    string
		Password []byte
		Name     string
		Created  int64
	}{email, pw, name, time.Now().Unix()})
	if err != nil {
		HTMLError(w, err.Error(), http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, "/register-success", http.StatusFound)
}

func HandleAuthn(w http.ResponseWriter, r *http.Request) {
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

func HandleLogout(w http.ResponseWriter, r *http.Request) {
	err := session.EndSession(w, r)
	if err != nil {
		HTMLError(w, err.Error(), http.StatusInternalServerError)
	}
	http.Redirect(w, r, "/", http.StatusFound)
}
