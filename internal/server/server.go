package server

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"

	"github.com/joho/godotenv"
	"github.com/odin-sofware/nyusu/internal/database"
	"github.com/odin-sofware/nyusu/internal/rss"
)

type Environment struct {
	DBUrl     string
	Engine    string
	Port      string
	Scrapper  int
	SecretKey []byte
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
	scrapper, err := strconv.Atoi(os.Getenv("SCRAPPER_TICK"))
	if err != nil {
		scrapper = 20
	}
	env := Environment{
		DBUrl:     os.Getenv("DB_URL"),
		Engine:    os.Getenv("DB_ENGINE"),
		Port:      fmt.Sprintf(":%s", os.Getenv("PORT")),
		SecretKey: []byte(os.Getenv("JWT_KEY")),
		Scrapper:  scrapper,
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

func (cfg *APIConfig) FetchPastFeeds(limit int) {
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
			rss, err := rss.DataFromFeed(url)
			if err != nil {
				fmt.Println((err.Error()))
			}
			err = cfg.DB.MarkFeedFetched(cfg.ctx, id)
			if err != nil {
				log.Println(err)
				return
			}
			for _, p := range rss.Channel.Items {
				t, err := ParseTime(p.Published)
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
		}(f.ID, f.Url)
	}
	wg.Wait()
	log.Println("Done")
}

func (cfg *APIConfig) FetchOneFeedSync(feedId int64, url string) {
	rss, err := rss.DataFromFeed(url)
	if err != nil {
		fmt.Println((err.Error()))
	}
	err = cfg.DB.MarkFeedFetched(cfg.ctx, feedId)
	if err != nil {
		log.Println(err)
		return
	}
	for _, p := range rss.Channel.Items {
		t, err := ParseTime(p.Published)
		if err != nil {
			log.Println(err)
		}
		_, err = cfg.DB.CreatePost(cfg.ctx, database.CreatePostParams{
			Title:       p.Title,
			Url:         p.Url,
			Description: sql.NullString{String: p.Description, Valid: true},
			FeedID:      feedId,
			PublishedAt: t.Unix(),
		})
		if err != nil {
			continue
		}
	}
}
