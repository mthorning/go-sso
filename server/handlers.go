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
	"path/filepath"
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
		ServeStaticPage(w, r, filepath.Clean(r.URL.Path), map[string]string{
			"Email": email,
			"Error": errorMessage,
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

	var dbUser types.DbUser
	doc.DataTo(&dbUser)

	if err = bcrypt.CompareHashAndPassword(dbUser.Password, []byte(password)); err != nil {
		sendError("Email or password incorrect")
		return
	}

	if err := session.SetSession(w, r, doc.Ref.ID); err != nil {
		HTMLError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

func HandleRegister(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		HTMLError(w, "Error reading form", http.StatusBadRequest)
		return
	}

	email := r.PostFormValue("email")
	password := r.PostFormValue("password")
	passwordAgain := r.PostFormValue("passwordAgain")
	name := r.PostFormValue("name")

	var sendError = func(errorMessage string) {
		ServeStaticPage(w, r, filepath.Clean(r.URL.Path), map[string]string{
			"Name":  name,
			"Email": email,
			"Error": errorMessage,
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
	if password != passwordAgain {
		sendError("Passwords do not match")
		return
	}

	unique, err := checkEmailUnique(w, email, "")
	if err != nil {
		HTMLError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !unique {
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
		Created  time.Time
	}{email, pw, name, time.Now()})
	if err != nil {
		HTMLError(w, err.Error(), http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, "/register-success", http.StatusFound)
}

func HandleAuthn(w http.ResponseWriter, r *http.Request) {
	// OLD, NEEDS FIXING
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

func HandleEdit(w http.ResponseWriter, r *http.Request) {
	sessionUser, err := getSessionUser(w, r)
	if err != nil {
		return
	}
	if err := r.ParseForm(); err != nil {
		HTMLError(w, "Error reading form", http.StatusBadRequest)
		return
	}

	email := r.PostFormValue("email")
	name := r.PostFormValue("name")

	var sendError = func(errorMessage string) {
		ServeStaticPage(w, r, filepath.Clean(r.URL.Path), map[string]string{
			"Email": email,
			"Name":  name,
			"Error": errorMessage,
		})
	}

	if email == "" {
		sendError("Email can't be blank")
		return
	}
	if name == "" {
		sendError("Name can't be blank")
		return
	}

	unique, err := checkEmailUnique(w, email, sessionUser.ID)
	if err != nil {
		HTMLError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !unique {
		sendError("Email address already taken")
		return
	}

	_, err = firestore.Users.Doc(sessionUser.ID).Update(context.Background(), []firestore.Update{
		{
			Path:  "Name",
			Value: name,
		},
		{
			Path:  "Email",
			Value: email,
		},
	})
	if err != nil {
		HTMLError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func HandleChpwd(w http.ResponseWriter, r *http.Request) {
	sessionUser, err := getSessionUser(w, r)
	if err != nil {
		return
	}

	if err := r.ParseForm(); err != nil {
		HTMLError(w, "Error reading form", http.StatusBadRequest)
		return
	}

	currentPassword := r.PostFormValue("currentPassword")
	password := r.PostFormValue("password")
	passwordAgain := r.PostFormValue("passwordAgain")

	var sendError = func(errorMessage string) {
		ServeStaticPage(w, r, filepath.Clean(r.URL.Path), map[string]string{
			"Error": errorMessage,
		})
	}
	if currentPassword == "" {
		sendError("Please enter your current password")
		return
	}
	if password == "" {
		sendError("Please enter a new password")
		return
	}
	if password != passwordAgain {
		sendError("Passwords do not match")
		return
	}

	ref := firestore.Users.Doc(sessionUser.ID)
	dsnap, err := ref.Get(context.Background())
	if err != nil {
		HTMLError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	dbUser := struct {
		Password []byte
	}{}
	err = dsnap.DataTo(&dbUser)
	if err != nil {
		HTMLError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := bcrypt.CompareHashAndPassword(dbUser.Password, []byte(currentPassword)); err != nil {
		sendError("Incorrect password")
		return
	}

	newPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		HTMLError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = ref.Update(context.Background(), []firestore.Update{
		{
			Path:  "Password",
			Value: newPassword,
		},
	})
	if err != nil {
		HTMLError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}
