package server

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/odin-sofware/nyusu/internal/database"
)

func (cfg *APIConfig) CreateFeed(w http.ResponseWriter, r *http.Request, user database.User) {
	var reqFeed *struct {
		Name string `json:"name,omitempty"`
		Url  string `json:"url,omitempty"`
	}
	err := json.NewDecoder(r.Body).Decode(&reqFeed)
	if err != nil {
		badRequestHandler(w)
		return
	}
	feed, err := cfg.DB.CreateFeed(cfg.ctx, database.CreateFeedParams{
		Name:   reqFeed.Name,
		Url:    reqFeed.Url,
		UserID: user.ID,
	})
	if err != nil {
		log.Print(err)
		internalServerErrorHandler(w)
		return
	}
	respondWithJSON(w, http.StatusCreated, feed)
}
