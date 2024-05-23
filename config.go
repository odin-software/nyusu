package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"math/rand/v2"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/odin-sofware/nyusu/internal/database"
)

type Environment struct {
	DBUrl  string
	Engine string
	Port   string
}

type Config struct {
	ctx context.Context
	DB  *database.Queries
	Env Environment
}

func NewConfig() Config {
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

	return Config{
		ctx: ctx,
		DB:  dbQueries,
		Env: env,
	}
}

func GetNewHash() string {
	r := strconv.FormatFloat(rand.Float64(), 'f', -1, 64)
	h := sha256.New()
	h.Write([]byte(r))

	return hex.EncodeToString(h.Sum((nil)))
}
