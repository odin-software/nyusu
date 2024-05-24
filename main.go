package main

import (
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
	"github.com/odin-sofware/nyusu/internal/server"
)

func main() {
	cfg := server.NewConfig()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/readiness", cfg.Readiness)
	mux.HandleFunc("GET /v1/err", cfg.Err)

	mux.HandleFunc("GET /v1/users", cfg.MiddlewareAuth(cfg.GetAuthUser))
	mux.HandleFunc("POST /v1/users", cfg.CreateUser)

	mux.HandleFunc("POST /v1/feeds", cfg.MiddlewareAuth(cfg.CreateFeed))

	log.Printf("server is listening at %s", cfg.Env.Port)
	log.Fatal(http.ListenAndServe(cfg.Env.Port, mux))
}
