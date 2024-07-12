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
	// mx := http.NewServeMux()
	mux.HandleFunc("GET /v1/readiness", cfg.Readiness)
	mux.HandleFunc("GET /v1/err", cfg.Err)

	// mux.HandleFunc("OPTIONS /v1/users/login", func(w http.ResponseWriter, r *http.Request) {
	// 	w.Header().Add("Access-Control-Allow-Origin", "*")
	// 	w.Header().Add("Access-Control-Allow-Methods", "*")
	// })
	// mux.HandleFunc("OPTIONS /v1/users", func(w http.ResponseWriter, r *http.Request) {
	// 	w.Header().Add("Access-Control-Allow-Origin", "*")
	// 	w.Header().Add("Access-Control-Allow-Methods", "*")
	// 	w.Header().Add("Access-Control-Allow-Headers", "*")
	// })

	mux.HandleFunc("POST /v1/users/login", cfg.LoginUser)
	mux.HandleFunc("POST /v1/users/register", cfg.RegisterUser)
	mux.HandleFunc("/v1/users", server.CORS(cfg.MiddlewareAuth(cfg.GetAuthUser)))

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

	go func() {
		for range ticker.C {
			cfg.FetchPastFeeds(5)
		}
	}()

	log.Printf("server is listening at %s", cfg.Env.Port)
	// log.Fatal(http.ListenAndServe(":4029", handler2))
	// log.Fatal(http.ListenAndServe(cfg.Env.Port, mx))
	log.Fatal(http.ListenAndServe(cfg.Env.Port, mux))
}
