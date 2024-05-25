package server

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"github.com/odin-sofware/nyusu/internal/database"
)

type Environment struct {
	DBUrl  string
	Engine string
	Port   string
}

type APIConfig struct {
	ctx context.Context
	DB  *database.Queries
	Env Environment
}

type AuthHandler func(http.ResponseWriter, *http.Request, database.User)

func NewConfig() APIConfig {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	env := Environment{
		DBUrl:  os.Getenv("DB_URL"),
		Engine: os.Getenv("DB_ENGINE"),
		Port:   fmt.Sprintf(":%s", os.Getenv("PORT")),
	}

	ctx := context.Background()
	db, err := sql.Open(env.Engine, env.DBUrl)
	if err != nil {
		log.Fatal(err)
	}
	dbQueries := database.New(db)

	return APIConfig{
		ctx: ctx,
		DB:  dbQueries,
		Env: env,
	}
}

func (cfg *APIConfig) MiddlewareAuth(handler AuthHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		key := strings.Split(header, " ")
		if key[0] != "ApiKey" || len(key) < 2 {
			unathorizedHandler(w)
			return
		}
		user, err := cfg.DB.GetUserByApiKey(cfg.ctx, key[1])
		if err != nil {
			log.Print(err)
			notFoundHandler(w)
			return
		}
		handler(w, r, user)
	}
}

func (cfg *APIConfig) Readiness(w http.ResponseWriter, r *http.Request) {
	payload := struct {
		Status string `json:"status"`
	}{
		Status: "ok",
	}
	respondWithJSON(w, http.StatusOK, payload)
}

func (cfg *APIConfig) Err(w http.ResponseWriter, r *http.Request) {
	internalServerErrorHandler(w)
}

func (cfg *APIConfig) TestXmlRes(limit int) {
	var wg sync.WaitGroup
	fs, err := cfg.DB.GetNextFeedsToFetch(cfg.ctx, int64(limit))
	if err != nil {
		log.Println(err)
		return
	}
	for _, f := range fs {
		wg.Add(1)
		go func(id int64, url string) {
			defer wg.Done()
			rss, err := DataFromFeed(url)
			if err != nil {
				fmt.Println((err.Error()))
			}
			for _, p := range rss.Channel.Items {
				t, err := time.Parse(time.RFC1123, p.Published)
				if err != nil {
					log.Println(err)
				}
				_, err = cfg.DB.CreatePost(cfg.ctx, database.CreatePostParams{
					Title:       p.Title,
					Url:         p.Url,
					Description: sql.NullString{String: p.Description, Valid: true},
					FeedID:      id,
					PublishedAt: t.Unix(),
				})
				if err != nil {
					continue
				}
			}
			println(rss.Channel.Title)
		}(f.ID, f.Url)
	}
	wg.Wait()
	log.Println("Done")
}
