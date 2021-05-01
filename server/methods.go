package server

import (
	"encoding/json"
	"github.com/mthorning/go-sso/jwt"
	"github.com/mthorning/go-sso/types"
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
			"code":  strconv.Itoa(code),
			"error": errStr,
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
