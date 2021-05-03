package server

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/mthorning/go-sso/firestore"
	"github.com/mthorning/go-sso/jwt"
	"github.com/mthorning/go-sso/session"
	"github.com/mthorning/go-sso/types"
	"google.golang.org/api/iterator"
	"html/template"
	"net/http"
	"path/filepath"
	"strconv"
)

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
		map[string]string{
			"Code":  strconv.Itoa(code),
			"Error": errStr,
		}); err != nil {
		return
	}
}

func JSONResponse(w http.ResponseWriter, response []byte) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

// Not sure about this yet
func getJWT(w http.ResponseWriter, user types.User) {
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

func getSessionUser(w http.ResponseWriter, r *http.Request) (types.SessionUser, error) {
	sessionUser, err := session.GetSession(w, r)
	if err != nil {
		if _, ok := err.(session.NoSessionError); ok {
			HTMLError(w, err.Error(), http.StatusForbidden)
		}
		return types.SessionUser{}, errors.New("Error getting session user")
	}
	return sessionUser, nil

}

func checkEmailUnique(w http.ResponseWriter, email, userID string) (bool, error) {
	query := firestore.Users.Where("Email", "==", email)
	iter := query.Documents(context.Background())
	defer iter.Stop()
	for {
		dsnap, err := iter.Next()
		if err == iterator.Done {
			return true, nil
		}
		if err != nil {
			return false, err
		}
		return dsnap.Ref.ID == userID, nil
	}
}
