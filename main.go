package main

import (
	"html/template"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/odin-sofware/nyusu/internal/server"
)

func main() {
	cfg := server.NewConfig()
	ticker := time.NewTicker(time.Duration(cfg.Env.Scrapper) * time.Second)

	// GET localhost:8888/v1/readiness, cfg.Readiness OK 200, Bad Request 400, Not Found 404
	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/readiness", cfg.Readiness)
	mux.HandleFunc("GET /v1/err", cfg.Err)

	// Page endpoints.
	mux.HandleFunc("GET /", cfg.GetHome)
	mux.HandleFunc("GET /login", func(w http.ResponseWriter, r *http.Request) {
		_, err := r.Cookie(server.SessionCookieName)
		if err == nil {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		t, err := template.ParseFiles("html/login.html")
		if err != nil {
			panic(err)
		}
		err = t.Execute(w, "")
		if err != nil {
			panic(err)
		}
	})

	mux.HandleFunc("POST /users/login", server.CORS(cfg.LoginUser))                   // post
	mux.HandleFunc("POST /users/logout", server.CORS(cfg.LogoutUser))                 // post
	mux.HandleFunc("POST /users/register", server.CORS(cfg.RegisterUser))             // post
	mux.HandleFunc("GET /v1/users", server.CORS(cfg.MiddlewareAuth(cfg.GetAuthUser))) // get

	mux.HandleFunc("GET /v1/feeds", server.CORS(cfg.GetAllFeeds))                     // get
	mux.HandleFunc("POST /v1/feeds", server.CORS(cfg.MiddlewareAuth(cfg.CreateFeed))) // post
	mux.HandleFunc("OPTIONS /v1/feeds", server.OPTIONS)                               // options

	mux.HandleFunc("GET /v1/feed_follows", server.CORS(cfg.MiddlewareAuth(cfg.GetFeedFollowsFromUser))) // get
	mux.HandleFunc("POST /v1/feed_follows", server.CORS(cfg.MiddlewareAuth(cfg.CreateFeedFollows)))     // post
	mux.HandleFunc("OPTIONS /v1/feed_follows", server.OPTIONS)                                          // options
	mux.HandleFunc("DELETE /v1/feed_follows/{feedFollowId}", cfg.DeleteFeedFollows)                     // delete

	mux.HandleFunc("DELETE /v1/posts/bookmarks/{postId}", server.CORS(cfg.MiddlewareAuth(cfg.UnbookmarkPost))) // delete
	mux.HandleFunc("POST /v1/posts/bookmarks/{postId}", server.CORS(cfg.MiddlewareAuth(cfg.BookmarkPost)))     // post
	mux.HandleFunc("OPTIONS /v1/posts/bookmarks/{postId}", server.OPTIONS)                                     // options
	mux.HandleFunc("GET /v1/posts/bookmarks", server.CORS(cfg.MiddlewareAuth(cfg.GetBookmarkedPosts)))         // get
	mux.HandleFunc("GET /v1/posts/{feedId}", server.CORS(cfg.MiddlewareAuth(cfg.GetPostByUsersAndFeed)))       // get
	mux.HandleFunc("GET /v1/posts", server.CORS(cfg.MiddlewareAuth(cfg.GetPostByUsers)))                       // get
	// server.TestRssParsing("https://frontendmasters.com/blog/feed/")
	// server.TestRssParsing("https://triss.dev/blog/rss.xml")

	go func() {
		for range ticker.C {
			cfg.FetchPastFeeds(5)
		}
	}()

	log.Printf("server is listening at %s", cfg.Env.Port)
	log.Fatal(http.ListenAndServe(cfg.Env.Port, mux))
}
