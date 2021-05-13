package main

import (
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/mthorning/go-sso/config"
	"github.com/mthorning/go-sso/firestore"
	"github.com/mthorning/go-sso/server"
	"github.com/mthorning/go-sso/types"
	"log"
	"net/http"
	"time"
)

type Config struct {
	Port int `default:"8080"`
}

var conf Config

func init() {
	config.SetConfig(&conf)
}

var routeConfig = server.RouteConfig{
	"/index": func(s *types.SessionUser) (interface{}, error) {
		d := struct {
			Admin bool
			Name  string
		}{}
		err := s.GetUser(&d)
		return d, err
	},
	"/edit": func(s *types.SessionUser) (interface{}, error) {
		d := struct {
			Name  string
			Email string
			Error string
		}{}
		err := s.GetUser(&d)
		return d, err
	},
	"/chpwd": func(s *types.SessionUser) (interface{}, error) {
		d := struct {
			Name  string
			Error string
		}{}
		err := s.GetUser(&d)
		return d, err
	},
	"/manage": func(s *types.SessionUser) (interface{}, error) {
		d := struct {
			Email string
		}{}
		err := s.GetUser(&d)
		if err != nil {
			return nil, err
		}

		query := firestore.Users.Where("Email", "!=", d.Email)
		docs, err := query.Documents(context.Background()).GetAll()
		if err != nil {
			return nil, err
		}

		var users []types.User
		for _, doc := range docs {
			var user types.User
			err := doc.DataTo(&user)
			if err != nil {
				return nil, err
			}
			user.ID = doc.Ref.ID
			users = append(users, user)
		}
		return users, nil
	},
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/login", server.HandleLogin).Methods("POST")
	r.HandleFunc("/register", server.HandleRegister).Methods("POST")
	r.HandleFunc("/authn", server.HandleAuthn).Methods("POST")
	r.HandleFunc("/logout", server.HandleLogout).Methods("POST")
	r.HandleFunc("/edit", server.HandleEdit).Methods("POST")
	r.HandleFunc("/chpwd", server.HandleChpwd).Methods("POST")

	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	r.HandleFunc("/login", server.NoAuthRoutes)
	r.HandleFunc("/register", server.NoAuthRoutes)
	r.HandleFunc("/register-success", server.NoAuthRoutes)

	authRoutes := server.AuthRoutes{
		Config: routeConfig,
	}
	r.PathPrefix("/").Handler(authRoutes)

	srv := &http.Server{
		Handler:      r,
		Addr:         fmt.Sprintf("127.0.0.1:%d", conf.Port),
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  30 * time.Second,
	}

	fmt.Println("Serving on port", conf.Port)
	log.Fatal(srv.ListenAndServe())
}
