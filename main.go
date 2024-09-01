package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"

	"github.com/a-h/templ"
	"github.com/joelramilison/timespent/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	
	dbUrl := loadEnv()
	db, err := sql.Open("postgres", dbUrl)
	if err != nil {
		panic(err)
	}
	cfg := &apiConfig{DB: database.New(db)}
	mux := http.NewServeMux()
	server := http.Server{Handler: mux}
	
	mux.Handle("GET /login", templ.Handler(login(nil)))
	mux.Handle("GET /register", templ.Handler(register()))
	mux.HandleFunc("POST /users", cfg.registerUserHandler)
	mux.HandleFunc("GET /{$}", cfg.middlewareAuth(cfg.appHandler))
	mux.HandleFunc("POST /login", cfg.loginHandler)
	mux.HandleFunc("GET /stopwatch", cfg.middlewareAuth(cfg.stopWatchHandler))
	mux.HandleFunc("POST /sessions/start", cfg.middlewareAuth(cfg.startSessionHandler))
	mux.Handle("POST /sessions/stop", templ.Handler(stopConfirmDialog("")))
	mux.HandleFunc("POST /sessions/confirm-stop", cfg.middlewareAuth(cfg.stopSessionHandler))
	mux.Handle("POST /sessions/abort-stop", templ.Handler(stopButton()))
	


	fmt.Println("Starting server ...")
	err = server.ListenAndServe()
	fmt.Printf("Server closed with error %v", err)

}

type apiConfig struct {
	DB *database.Queries
	
}

func loadEnv() (string) {

	err := godotenv.Load()
	if err != nil {
		panic(err)
	}
	dbUrl := os.Getenv("DB_CONNECTION_STRING")
	if dbUrl == "" {
		panic("empty dbUrl from environment variables")
	}
	return dbUrl
}