package main

import (
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/odin-sofware/nyusu/internal/server"
	"github.com/rs/cors"
)

func main() {
	cfg := server.NewConfig()
	ticker := time.NewTicker(time.Duration(cfg.Env.Scrapper) * time.Second)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/readiness", cfg.Readiness)
	mux.HandleFunc("GET /v1/err", cfg.Err)

	mux.HandleFunc("POST /v1/users/login", cfg.LoginUser)
	mux.HandleFunc("POST /v1/users/register", cfg.RegisterUser)
	mux.HandleFunc("GET /v1/users", cfg.MiddlewareAuth(cfg.GetAuthUser))

	mux.HandleFunc("GET /v1/feeds", cfg.GetAllFeeds)
	mux.HandleFunc("POST /v1/feeds", cfg.MiddlewareAuth(cfg.CreateFeed))

	mux.HandleFunc("GET /v1/feed_follows", cfg.MiddlewareAuth(cfg.GetFeedFollowsFromUser))
	mux.HandleFunc("POST /v1/feed_follows", cfg.MiddlewareAuth(cfg.CreateFeedFollows))
	mux.HandleFunc("DELETE /v1/feed_follows/{feedFollowId}", cfg.DeleteFeedFollows)

	mux.HandleFunc("DELETE /v1/posts/bookmarks/{postId}", cfg.MiddlewareAuth(cfg.UnbookmarkPost))
	mux.HandleFunc("POST /v1/posts/bookmarks/{postId}", cfg.MiddlewareAuth(cfg.BookmarkPost))
	mux.HandleFunc("GET /v1/posts/bookmarks", cfg.MiddlewareAuth(cfg.GetBookmarkedPosts))
	mux.HandleFunc("GET /v1/posts/{feedId}", cfg.MiddlewareAuth(cfg.GetPostByUsersAndFeed))
	mux.HandleFunc("GET /v1/posts", cfg.MiddlewareAuth(cfg.GetPostByUsers))
	// server.Basic()

	handler := cors.Default().Handler(mux)

	go func() {
		for range ticker.C {
			cfg.FetchPastFeeds(5)
		}
	}()

	log.Printf("server is listening at %s", cfg.Env.Port)
	log.Fatal(http.ListenAndServe(cfg.Env.Port, handler))
}
