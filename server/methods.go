package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/mthorning/go-sso/firestore"
	"github.com/mthorning/go-sso/jwt"
	"github.com/mthorning/go-sso/session"
	"github.com/mthorning/go-sso/types"
	"google.golang.org/api/iterator"
	"html/template"
	"net/http"
	"path/filepath"
	"runtime"
	"strconv"
	"time"
)

func trace() string {
	pc := make([]uintptr, 15)
	n := runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc[:n])

	// skip first frame as that is the func which called this
	_, _ = frames.Next()
	trace := ""
	for {
		frame, more := frames.Next()
		if !more {
			break
		}
		trace = fmt.Sprintf("%s\n%s:%d", trace, frame.File, frame.Line)
	}
	return trace
}

func JSONError(w http.ResponseWriter, err string, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"message": err})
}

func HTMLError(w http.ResponseWriter, r *http.Request, errStr string, code int) {
	lp := filepath.Join("templates", "layout.html")
	ep := filepath.Join("templates", "error.html")

	tmpl, err := makeTemplate(w, r, lp, ep)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error in HTMLError ParseFiles: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	origin := trace()

	if err := tmpl.ExecuteTemplate(w, "layout",
		map[string]string{
			"Code":   strconv.Itoa(code),
			"Error":  errStr,
			"Origin": origin,
		}); err != nil {
		http.Error(w, fmt.Sprintf("Error in HTMLError ExecuteTemplate: %s", err.Error()), http.StatusInternalServerError)
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
			HTMLError(w, r, err.Error(), http.StatusForbidden)
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

func makeTemplate(w http.ResponseWriter, r *http.Request, files ...string) (*template.Template, error) {

	funcMap := template.FuncMap{
		"many": func(s ...string) []string {
			return s
		},
		"isLoggedIn": func(_ ...string) bool {
			_, err := session.GetSession(w, r)
			return err == nil
		},
		"yesNo": func(x bool) string {
			if x {
				return "Yes"
			}
			return "No"
		},
		"dateTime": func(t time.Time) string {
			return t.Format("2006-01-02 15:04:05")
		},
	}

	return template.New("page").Funcs(funcMap).ParseFiles(files...)
}
