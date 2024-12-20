package main

import (
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/odin-software/nyusu/internal/server"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

func main() {
	cfg := server.NewConfig()
	ticker := time.NewTicker(time.Duration(cfg.Env.Scrapper) * time.Second)
	fs := http.FileServer(http.Dir("./static"))

	mux := http.NewServeMux()
	mux.Handle("GET /static/", http.StripPrefix("/static/", fs))

	// Page endpoints.
	mux.HandleFunc("GET /", cfg.GetHome)
	mux.HandleFunc("GET /login", cfg.GetLogin)
	mux.HandleFunc("GET /register", cfg.GetRegister)
	mux.HandleFunc("GET /add", cfg.GetAddFeed)
	mux.HandleFunc("GET /feeds", cfg.GetAllFeeds)
	mux.HandleFunc("GET /feeds/{feedId}", cfg.GetFeedPosts)

	// Action endpoints.
	mux.HandleFunc("POST /users/login", cfg.LoginUser)
	mux.HandleFunc("POST /users/logout", cfg.LogoutUser)
	mux.HandleFunc("POST /users/register", cfg.RegisterUser)
	mux.HandleFunc("POST /feed", cfg.CreateFeed)

	mux.HandleFunc("GET /v1/feeds", server.CORS(cfg.GetAllFeeds2))                                      // get
	mux.HandleFunc("GET /v1/feed_follows", server.CORS(cfg.MiddlewareAuth(cfg.GetFeedFollowsFromUser))) // get
	mux.HandleFunc("DELETE /v1/feed_follows/{feedFollowId}", cfg.DeleteFeedFollows)                     // delete

	mux.HandleFunc("DELETE /v1/posts/bookmarks/{postId}", server.CORS(cfg.MiddlewareAuth(cfg.UnbookmarkPost))) // delete
	mux.HandleFunc("POST /v1/posts/bookmarks/{postId}", server.CORS(cfg.MiddlewareAuth(cfg.BookmarkPost)))     // post
	mux.HandleFunc("GET /v1/posts/bookmarks", server.CORS(cfg.MiddlewareAuth(cfg.GetBookmarkedPosts)))         // get

	go func() {
		for range ticker.C {
			cfg.FetchPastFeeds(5)
		}
	}()

	log.Printf("server is listening at %s", cfg.Env.Port)
	log.Fatal(http.ListenAndServe(cfg.Env.Port, mux))
}
