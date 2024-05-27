package main

import (
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/odin-sofware/nyusu/internal/server"
)

func main() {
	cfg := server.NewConfig()
	ticker := time.NewTicker(time.Duration(cfg.Env.Scrapper) * time.Second)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/readiness", cfg.Readiness)
	mux.HandleFunc("GET /v1/err", cfg.Err)

	mux.HandleFunc("GET /v1/users", cfg.MiddlewareAuth(cfg.GetAuthUser))
	mux.HandleFunc("POST /v1/users", cfg.CreateUser)

	mux.HandleFunc("GET /v1/feeds", cfg.GetAllFeeds)
	mux.HandleFunc("POST /v1/feeds", cfg.MiddlewareAuth(cfg.CreateFeed))

	mux.HandleFunc("GET /v1/feed_follows", cfg.MiddlewareAuth(cfg.GetFeedFollowsFromUser))
	mux.HandleFunc("POST /v1/feed_follows", cfg.MiddlewareAuth(cfg.CreateFeedFollows))
	mux.HandleFunc("DELETE /v1/feed_follows/{feedFollowId}", cfg.DeleteFeedFollows)

	mux.HandleFunc("GET /v1/posts/{feedId}", cfg.MiddlewareAuth(cfg.GetPostByUsersAndFeed))
	mux.HandleFunc("GET /v1/posts", cfg.MiddlewareAuth(cfg.GetPostByUsers))

	go func() {
		for range ticker.C {
			cfg.FetchPastFeeds(5)
		}
	}()

	log.Printf("server is listening at %s", cfg.Env.Port)
	log.Fatal(http.ListenAndServe(cfg.Env.Port, mux))
}
