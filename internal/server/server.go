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
	"github.com/odin-software/nyusu/internal/database"
	"github.com/odin-software/nyusu/internal/rss"
)

type Environment struct {
	DBUrl         string
	Engine        string
	Port          string
	Scrapper      int
	SecretKey     []byte
	Environment   string
	ProductionURL string
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
		log.Println("Error loading .env file")
	}
	scrapper, err := strconv.Atoi(os.Getenv("SCRAPPER_TICK"))
	if err != nil {
		scrapper = 20
	}
	environment := os.Getenv("ENVIRONMENT")
	if environment == "" {
		environment = "development"
	}

	productionURL := os.Getenv("PRODUCTION_URL")
	if productionURL == "" {
		productionURL = "https://nyusu.do"
	}

	env := Environment{
		DBUrl:         os.Getenv("DB_URL"),
		Engine:        os.Getenv("DB_ENGINE"),
		Port:          fmt.Sprintf(":%s", os.Getenv("PORT")),
		SecretKey:     []byte(os.Getenv("JWT_KEY")),
		Scrapper:      scrapper,
		Environment:   environment,
		ProductionURL: productionURL,
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
				log.Printf("Failed to fetch RSS feed (ID: %d, URL: %s): %s", id, url, err.Error())
				return // Skip processing if RSS fetch failed
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
				author := p.Author
				if author == "" {
					author = p.Creator
				}
				_, err = cfg.DB.CreatePost(cfg.ctx, database.CreatePostParams{
					Title:       p.Title,
					Url:         p.Url,
					Author:      author,
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
	log.Println("Finished fetching feeds")
}

func (cfg *APIConfig) FetchOneFeedSync(feedId int64, url string) {
	rss, err := rss.DataFromFeed(url)
	if err != nil {
		log.Printf("Failed to fetch RSS feed (ID: %d, URL: %s): %s", feedId, url, err.Error())
		return // Skip processing if RSS fetch failed
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
		author := p.Author
		if author == "" {
			author = p.Creator
		}
		_, err = cfg.DB.CreatePost(cfg.ctx, database.CreatePostParams{
			Title:       p.Title,
			Url:         p.Url,
			Author:      author,
			Description: sql.NullString{String: p.Description, Valid: true},
			FeedID:      feedId,
			PublishedAt: t.Unix(),
		})
		if err != nil {
			continue
		}
	}
}
