package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"github.com/odin-sofware/nyusu/internal/database"
)

func main() {
	config := NewConfig()

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
	mux.HandleFunc("GET /v1/users", func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		key := strings.Split(header, " ")
		if key[0] != "ApiKey" || len(key) < 2 {
			UnathorizedHandler(w)
			return
		}
		user, err := config.DB.GetUserByApiKey(config.ctx, key[1])
		if err != nil {
			log.Print(err)
			NotFoundHandler(w)
			return
		}
		respondWithJSON(w, 201, user)
	})
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
		user, err := config.DB.CreateUser(config.ctx, database.CreateUserParams{
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
	log.Printf("server is listening at %s", config.Env.Port)
	log.Fatal(http.ListenAndServe(config.Env.Port, mux))
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
