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

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/joho/godotenv"
	hubclient "github.com/odin-software/hub-client"
	"github.com/odin-software/nyusu/internal/database"
	"github.com/odin-software/nyusu/internal/rss"
	"github.com/pressly/goose/v3"
	"golang.org/x/oauth2"
)

type Environment struct {
	DBUrl            string
	Port             string
	Scrapper         int
	Environment      string
	ProductionURL    string
	OIDCIssuerURL    string
	OIDCClientID     string
	OIDCClientSecret string
	OIDCRedirectURL string
}

type Branding struct {
	Name           string
	PrimaryColor   string
	SecondaryColor string
	TertiaryColor  string
	QuartaryColor  string
	CardBg         string
	BorderColor    string
}

type APIConfig struct {
	ctx          context.Context
	DB           *database.Queries
	Env          Environment
	Branding     Branding
	OIDCProvider *oidc.Provider
	OAuth2Config oauth2.Config
}

type AuthHandler func(http.ResponseWriter, *http.Request, database.User)

func NewConfig() APIConfig {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}

	// Fetch config from Hub. Falls back to local env vars if Hub is
	// not configured or unreachable.
	cfg := hubclient.Load()

	scrapper, err := strconv.Atoi(cfg.Get("SCRAPPER_TICK", os.Getenv("SCRAPPER_TICK")))
	if err != nil {
		scrapper = 20
	}

	environment := cfg.Get("ENVIRONMENT", os.Getenv("ENVIRONMENT"))
	if environment == "" {
		environment = "development"
	}

	productionURL := cfg.Get("PRODUCTION_URL", os.Getenv("PRODUCTION_URL"))
	if productionURL == "" {
		productionURL = "https://nyusu.odin.do"
	}

	port := cfg.Get("PORT", os.Getenv("PORT"))

	env := Environment{
		DBUrl:            cfg.Get("DB_URL", os.Getenv("DB_URL")),
		Port:             fmt.Sprintf(":%s", port),
		Scrapper:         scrapper,
		Environment:      environment,
		ProductionURL:    productionURL,
		OIDCIssuerURL:    cfg.Get("OIDC_ISSUER_URL", os.Getenv("OIDC_ISSUER_URL")),
		OIDCClientID:     cfg.Get("OIDC_CLIENT_ID", os.Getenv("OIDC_CLIENT_ID")),
		OIDCClientSecret: cfg.Get("OIDC_CLIENT_SECRET", os.Getenv("OIDC_CLIENT_SECRET")),
		OIDCRedirectURL:  cfg.Get("OIDC_REDIRECT_URL", os.Getenv("OIDC_REDIRECT_URL")),
	}

	ctx := context.Background()
	db, err := sql.Open("pgx", env.DBUrl)
	if err != nil {
		log.Fatal(err)
	}

	// Run database migrations
	log.Println("Running database migrations...")
	if err := runMigrations(db); err != nil {
		log.Printf("Warning: Migration error: %v", err)
	} else {
		log.Println("Migrations completed successfully")
	}

	dbQueries := database.New(db)

	// Initialize OIDC provider
	provider, err := oidc.NewProvider(ctx, env.OIDCIssuerURL)
	if err != nil {
		log.Fatalf("Failed to initialize OIDC provider: %v", err)
	}

	oauth2Config := oauth2.Config{
		ClientID:     env.OIDCClientID,
		ClientSecret: env.OIDCClientSecret,
		RedirectURL:  env.OIDCRedirectURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	// Global branding: Epsilon had a separate global config section. Hub
	// doesn't (yet), so we pass an empty map — buildBranding falls back
	// to Nyusu's defaults for every key.
	branding := buildBranding(map[string]string{})

	return APIConfig{
		ctx:          ctx,
		DB:           dbQueries,
		Env:          env,
		Branding:     branding,
		OIDCProvider: provider,
		OAuth2Config: oauth2Config,
	}
}

func runMigrations(db *sql.DB) error {
	goose.SetBaseFS(nil)
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	return goose.Up(db, "sql/schema")
}

// configValue and fetchEpsilonConfig were here — removed in the Hub
// migration (v0.4.3). Config is now fetched via hubclient.Load() at
// the top of NewConfig().

// buildBranding extracts branding values from the global config map,
// falling back to Nyusu's defaults for any missing key.
func buildBranding(global map[string]string) Branding {
	branding := Branding{
		Name:           "Odin Software",
		PrimaryColor:   "#DDE61F",
		SecondaryColor: "#1A5632",
		TertiaryColor:  "#0F1822",
		QuartaryColor:  "#D6D7D5",
		CardBg:         "#1a2332",
		BorderColor:    "#2a3544",
	}

	if global == nil {
		return branding
	}

	if v, ok := global["BRAND_NAME"]; ok && v != "" {
		branding.Name = v
	}
	if v, ok := global["BRAND_PRIMARY_COLOR"]; ok && v != "" {
		branding.PrimaryColor = v
	}
	if v, ok := global["BRAND_SECONDARY_COLOR"]; ok && v != "" {
		branding.SecondaryColor = v
	}
	if v, ok := global["BRAND_TERTIARY_COLOR"]; ok && v != "" {
		branding.TertiaryColor = v
	}
	if v, ok := global["BRAND_QUARTARY_COLOR"]; ok && v != "" {
		branding.QuartaryColor = v
	}
	if v, ok := global["BRAND_CARD_BG"]; ok && v != "" {
		branding.CardBg = v
	}
	if v, ok := global["BRAND_BORDER_COLOR"]; ok && v != "" {
		branding.BorderColor = v
	}

	return branding
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
	fs, err := cfg.DB.GetNextFeedsToFetch(cfg.ctx, int32(limit))
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
					PublishedAt: t,
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
			PublishedAt: t,
		})
		if err != nil {
			continue
		}
	}
}
