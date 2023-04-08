package main

import (
	"appurl/handlers"
	"appurl/repository"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/joho/godotenv/autoload"
	_ "github.com/lib/pq"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatalf("err loading: %v", err)
	}

	db, err := repository.InitPostgresDb(repository.PsqlConfig{
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

	rep := repository.NewInquirysRepository(db)
	myhandler := handlers.NewUseRepository(rep)

	router := mux.NewRouter()
	router.HandleFunc("/{url_index}", myhandler.RedirectShortUrl).Methods("GET")
	router.HandleFunc("/take_larg_url", myhandler.CreateShortUrl).Methods("POST")

	log.Fatal(http.ListenAndServe(":8000", router))

}
