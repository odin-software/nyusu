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

	mux.HandleFunc("/v1/users/login", server.CORS(cfg.LoginUser))                 // post
	mux.HandleFunc("/v1/users/register", server.CORS(cfg.RegisterUser))           // post
	mux.HandleFunc("/v1/users", server.CORS(cfg.MiddlewareAuth(cfg.GetAuthUser))) // get

	mux.HandleFunc("/v1/feeds", server.CORS(cfg.GetAllFeeds))                    // get
	mux.HandleFunc("/v1/feeds", server.CORS(cfg.MiddlewareAuth(cfg.CreateFeed))) // post

	mux.HandleFunc("GET /v1/feed_follows", cfg.MiddlewareAuth(cfg.GetFeedFollowsFromUser)) // get
	mux.HandleFunc("POST /v1/feed_follows", cfg.MiddlewareAuth(cfg.CreateFeedFollows))     // post
	mux.HandleFunc("DELETE /v1/feed_follows/{feedFollowId}", cfg.DeleteFeedFollows)        // delete

	mux.HandleFunc("DELETE /v1/posts/bookmarks/{postId}", cfg.MiddlewareAuth(cfg.UnbookmarkPost)) // delete
	mux.HandleFunc("POST /v1/posts/bookmarks/{postId}", cfg.MiddlewareAuth(cfg.BookmarkPost))     // post
	mux.HandleFunc("GET /v1/posts/bookmarks", cfg.MiddlewareAuth(cfg.GetBookmarkedPosts))         // get
	mux.HandleFunc("GET /v1/posts/{feedId}", cfg.MiddlewareAuth(cfg.GetPostByUsersAndFeed))       // get
	mux.HandleFunc("/v1/posts", server.CORS(cfg.MiddlewareAuth(cfg.GetPostByUsers)))              // get

	go func() {
		for range ticker.C {
			cfg.FetchPastFeeds(5)
		}
	}()

	log.Printf("server is listening at %s", cfg.Env.Port)
	log.Fatal(http.ListenAndServe(cfg.Env.Port, mux))
}
