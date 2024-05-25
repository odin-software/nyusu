package server

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

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

func (cfg *APIConfig) TestXmlRes(w http.ResponseWriter, r *http.Request) {
	DataFromFeed("https://blog.boot.dev/index.xml")
	respondOk(w)
}
