package server

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/odin-sofware/nyusu/internal/database"
)

func (cfg *APIConfig) GetAuthUser(w http.ResponseWriter, r *http.Request, user database.User) {
	respondWithJSON(w, http.StatusOK, user)
}

func (cfg *APIConfig) CreateUser(w http.ResponseWriter, r *http.Request) {
	var reqUser *struct {
		Name string `json:"name"`
	}
	err := json.NewDecoder(r.Body).Decode(&reqUser)
	if err != nil {
		badRequestHandler(w)
		return
	}
	key := GetNewHash()
	user, err := cfg.DB.CreateUser(cfg.ctx, database.CreateUserParams{
		Name:   reqUser.Name,
		ApiKey: key,
	})
	if err != nil {
		log.Print(err)
		internalServerErrorHandler(w)
		return
	}
	respondWithJSON(w, http.StatusCreated, user)
}
