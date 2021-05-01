package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/mthorning/go-sso/config"
	"github.com/mthorning/go-sso/server"
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

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/login", server.HandleLogin).Methods("POST")
	r.HandleFunc("/register", server.HandleRegister).Methods("POST")
	r.HandleFunc("/authn", server.HandleAuthn).Methods("POST")

	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	r.HandleFunc("/login", server.NoAuthRoutes)
	r.HandleFunc("/register", server.NoAuthRoutes)
	r.HandleFunc("/register-success", server.NoAuthRoutes)
	r.PathPrefix("/").Handler(server.AuthRoutes{})

	srv := &http.Server{
		Handler:      r,
		Addr:         fmt.Sprintf("127.0.0.1:%d", conf.Port),
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  30 * time.Second,
	}

	fmt.Println("Serving on port", conf.Port)
	log.Fatal(srv.ListenAndServe())
}
