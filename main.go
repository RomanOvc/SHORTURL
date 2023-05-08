package main

import (
	"appurl/handlers"
	"appurl/repository"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatalf("err loading: %v", err)
	}

	dbPostgres, err := repository.InitPostgresDb(repository.PsqlConfig{
		Host:     os.Getenv("host"),
		Port:     os.Getenv("port"),
		Username: os.Getenv("username"),
		Dbname:   os.Getenv("dbname"),
		Password: os.Getenv("pass"),
		Sslmode:  os.Getenv("sslmode"),
	})
	if err != nil {
		log.Fatal(err)
	}

	dbRedis := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	redisClient := repository.NewRedisreposiory(dbRedis)

	authrep := repository.NewAuthInquirysRepository(dbPostgres)
	authHandler := handlers.NewAuthInquirysRepository(authrep, redisClient)
	rep := repository.NewInquirysRepository(dbPostgres)
	myhandler := handlers.NewUseRepository(rep)

	router := mux.NewRouter()

	// router.HandleFunc("/{url_index}", myhandler.RedirectShortUrl).Methods("GET")
	//  router.HandleFunc("/take_larg_urls", myhandler.CreateShortUrl).Methods("POST")

	// FIXME авторизация на использование не нужна, добавить запись с юзер агентом в базу
	router.Handle("/{url_index}", authHandler.IsAuth(myhandler.RedirectShortUrl)).Methods("GET")
	router.Handle("/take_larg_url", authHandler.IsAuth(myhandler.CreateShortUrl)).Methods("POST")

	// create handlers
	// FIXME auth block
	router.HandleFunc("/create_user", authHandler.CreateUserH).Methods("POST")
	router.HandleFunc("/create_user/activate/{uuid}", authHandler.EmailActivateH).Methods("GET")

	// authentication handler (auth)

	// Регистрация 		   POST /auth
	// Авторизация 		   PUT /auth
	// Рефреш 	   		   POST /auth/refresh
	// Подтверждение почты GET /auth/confirm/{uuid}
	router.HandleFunc("/authentication", authHandler.AuthentificateUserH).Methods("POST")
	router.HandleFunc("/authentication/refresh_token", authHandler.RefreshTokenH).Methods("POST")
	router.HandleFunc("/authentication/forgot_pass", authHandler.ForgotPasswordH).Methods("POST") //сменить на PATCH

	// router.HandleFunc("/update_status", authHandler.UpdateUserH).Methods("POST")
	server := http.Server{
		Addr:              ":8000",
		Handler:           router,
		ReadTimeout:       time.Second * 10,
		WriteTimeout:      time.Second * 10,
		ReadHeaderTimeout: time.Second * 10,
		IdleTimeout:       time.Second * 10,
	}
	log.Fatal(server.ListenAndServe())

}
