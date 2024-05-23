package main

import (
	"encoding/json"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
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
	mux.HandleFunc("POST /v1/users", func(w http.ResponseWriter, r *http.Request) {
		var reqUser *struct {
			Name string `json:"name"`
		}
		err := json.NewDecoder(r.Body).Decode(&reqUser)
		if err != nil {
			BadRequestErrorHandler(w, "Malformed create user request")
			return
		}
		user, err := config.DB.CreateUser(config.ctx, reqUser.Name)
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
