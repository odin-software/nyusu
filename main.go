package main

import (
	"encoding/json"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
	"github.com/odin-sofware/nyusu/internal/database"
)

func main() {
	cfg := NewConfig()

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/readiness", func(w http.ResponseWriter, r *http.Request) {
		payload := struct {
			status string
		}{
			status: "ok",
		}
		respondWithJSON(w, 200, payload)
	})
	mux.HandleFunc("/v1/err", func(w http.ResponseWriter, r *http.Request) {
		InternalServerErrorHandler(w)
	})
	mux.HandleFunc("GET /v1/users", cfg.middlewareAuth(func(w http.ResponseWriter, r *http.Request, user database.User) {
		respondWithJSON(w, 200, user)
	}))
	mux.HandleFunc("POST /v1/users", func(w http.ResponseWriter, r *http.Request) {
		var reqUser *struct {
			Name string `json:"name"`
		}
		err := json.NewDecoder(r.Body).Decode(&reqUser)
		if err != nil {
			BadRequestHandler(w, "Malformed create user request")
			return
		}
		key := GetNewHash()
		user, err := cfg.DB.CreateUser(cfg.ctx, database.CreateUserParams{
			Name:   reqUser.Name,
			ApiKey: key,
		})
		if err != nil {
			log.Print(err)
			InternalServerErrorHandler(w)
			return
		}
		respondWithJSON(w, 201, user)
	})
	log.Printf("server is listening at %s", cfg.Env.Port)
	log.Fatal(http.ListenAndServe(cfg.Env.Port, mux))
}

func respondWithJSON(w http.ResponseWriter, status int, payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		InternalServerErrorHandler(w)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(data)
}
