package main

import (
	"appurl/crontasks"
	"appurl/handlers"
	"appurl/middlewars"
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

	// инисиальизасия postgres
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

	//
	dbRedisTable0 := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	log.Println("start server")
	redisClient := repository.NewRedisReposiory(dbRedisTable0)
	authrep := repository.NewAuthInquirysRepository(dbPostgres)
	authHandler := handlers.NewAuthInquirysRepository(authrep, redisClient)
	rep := repository.NewInquirysRepository(dbPostgres)
	myhandler := handlers.NewUseRepository(rep)

	router := mux.NewRouter()

	//TODO cron task
	go crontasks.RunCronJob(dbPostgres)

	// isAuth := middlewars.NewAuthInquirysRepository(authrep, redisClient)
	isAuth := middlewars.IsAuth(authrep, redisClient)

	//
	router.HandleFunc("/{url_index}", myhandler.RedirectShortUrl).Methods("GET")
	// Аворизаия нужна
	router.Handle("/take_larg_url", isAuth(myhandler.CreateShortUrl)).Methods("POST")
	router.Handle("/user/all_user_urls", isAuth(myhandler.AllUsersUrls)).Methods("GET")

	// статистика посещении по юрл
	router.Handle("/user/statistic/{url_index}", isAuth(myhandler.VisitOnUrlH)).Methods("GET")
	router.Handle("/user/statistic/{url_index}/count_visit", isAuth(myhandler.CountVisitH)).Methods("GET")

	// auth block
	// Регистрация 		   POST /auth
	// Авторизация 		   PUT /auth
	// Рефреш токен 	   POST /auth/refresh
	// Подтверждение почты GET /auth/confirm/{uuid}
	router.HandleFunc("/auth", authHandler.CreateUserH).Methods("POST")
	router.HandleFunc("/auth", authHandler.AuthentificateUserH).Methods("PUT")
	router.HandleFunc("/auth/refresh", authHandler.RefreshTokenH).Methods("POST")
	router.HandleFunc("/auth/confirm/{uid}", authHandler.EmailActivateH).Methods("GET")
	router.HandleFunc("/auth/forgotpass", authHandler.ForgotPasswordH).Methods("POST") //сменить на PATCH
	router.HandleFunc("/auth/resetpass/{resetToken}", authHandler.ResetPassH).Methods("POST")

	router.HandleFunc("/auth/repeatconfirm", authHandler.RepeatEmailActivateH).Methods("POST")
	server := http.Server{
		Addr:              ":8001",
		Handler:           router,
		ReadTimeout:       time.Second * 10,
		WriteTimeout:      time.Second * 10,
		ReadHeaderTimeout: time.Second * 10,
		IdleTimeout:       time.Second * 10,
	}
	log.Fatal(server.ListenAndServe())

}
