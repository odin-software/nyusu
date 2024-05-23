package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

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
