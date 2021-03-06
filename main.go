package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"todo-app/database"
	"todo-app/handlers"
	"todo-app/middleware"
	"todo-app/util"
	"todo-app/util/auth"

	"github.com/go-redis/redis/v7"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetFormatter(&log.TextFormatter{ForceColors: true, FullTimestamp: true})
	log.SetReportCaller(true)
}

func NewRedisDB(host, port, password string) (*redis.Client, error) {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     host + ":" + port,
		Password: password,
		DB:       0,
	})

	// if we can't talk to redis, fail fast
	if _, err := redisClient.Ping().Result(); err != nil {
		return nil, errors.Wrap(err, "Could not connect to redis db0")
	}
	// ret := &Store{c: c}
	// return ret, nil
	return redisClient, nil
}

func main() {
	log.Info("Starting Up Todolist API")
	conf, err := getConfig()
	if err != nil {
		log.Fatalf("Failed to read config: %v", err)
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	redisClient, err := NewRedisDB(conf.RedisHost, conf.RedisPort, conf.RedisPassword)
	if err != nil {
		log.Fatalf("Failed to create redis store: %v", err)
	}

	token := auth.NewToken()
	rdAuth := auth.NewAuth(redisClient)

	store, err := database.New()
	if err != nil {
		log.Fatalf("Failed to create store: %v", err)
	}

	handler := handlers.NewHandler(store, token, rdAuth)

	serveMux := mux.NewRouter()
	serveMux.HandleFunc("/signup", handler.Signup).Methods("POST")
	serveMux.HandleFunc("/signin", handler.Signin).Methods("POST")
	serveMux.HandleFunc("/logout", handler.Logout).Methods("GET", "POST")
	serveMux.HandleFunc("/tokens", handler.Refresh).Methods("POST")
	serveMux.HandleFunc("/ws", handler.WSEndpoint)

	tasksRouter := serveMux.PathPrefix("/tasks").Subrouter()
	tasksRouter.Use(middleware.AuthMiddleware)
	tasksRouter.HandleFunc("", handler.CreateTask).Methods("POST")
	tasksRouter.HandleFunc("", handler.GetTasks).Queries("completed", "{completed:true|false}").Methods("GET")
	tasksRouter.HandleFunc("", handler.GetTasks).Queries("priority", "{priority:[1-3]}").Methods("GET")
	tasksRouter.HandleFunc("", handler.GetTasks).Methods("GET")
	tasksRouter.HandleFunc("/{id:[0-9]+}", handler.GetTask).Methods("GET")
	tasksRouter.HandleFunc("/{id:[0-9]+}", handler.UpdateTask).Methods(http.MethodPatch)
	tasksRouter.HandleFunc("/{id:[0-9]+}", handler.DeleteTask).Methods(http.MethodDelete)
	tasksRouter.HandleFunc("/{idTask:[0-9]+}/{idUser:[0-9]+}", handler.AddUserToTask).Methods(http.MethodPost)
	tasksRouter.HandleFunc("/{idTask:[0-9]+}/{idUser:[0-9]+}", handler.RemoveUserFromTask).Methods(http.MethodDelete)

	go handlers.Reader()

	s := &http.Server{
		Addr:         ":8080",
		Handler:      serveMux,
		WriteTimeout: 2 * time.Second,
	}

	go func() {
		log.Fatal(s.ListenAndServe())
	}()

	sig := <-sigs
	log.Warningf("Received signal %s, Terminating gracefully", sig)
	tc, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	s.Shutdown(tc)
}

func getConfig() (util.EnvVariables, error) {
	if err := godotenv.Load(); err != nil {
		return util.EnvVariables{}, errors.Wrap(err, "failed to load env file")
	}

	conf, err := util.GetConfig()
	if err != nil {
		return util.EnvVariables{}, errors.Wrap(err, "failed to load config vars")
	}

	return conf, nil
}
