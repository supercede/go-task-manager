package main

import (
	"log"
	"net/http"
	"time"
	"todo-app/database"
	"todo-app/handlers"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("failed to load env file", err)
	}

	store, err := database.New()
	if err != nil {
		log.Fatalf("Failed to create store: %v", err)
	}

	handler := handlers.NewHandler(store)

	route := mux.NewRouter()
	route.HandleFunc("/signup", handler.Signup).Methods("POST")
	route.HandleFunc("/signin", handler.Signin).Methods("POST")

	s := &http.Server{
		Addr:         ":8080",
		Handler:      route,
		WriteTimeout: 2 * time.Second,
	}

	log.Fatal(s.ListenAndServe())
}

// func getConfig() (util.EnvVariables, error) {
// 	if err := godotenv.Load(); err != nil {
// 		return util.EnvVariables{}, errors.Wrap(err, "failed to load env file")
// 	}

// 	conf, err := util.GetConfig()
// 	if err != nil {
// 		return util.EnvVariables{}, errors.Wrap(err, "failed to load config vars")
// 	}

// 	return conf, nil
// }
