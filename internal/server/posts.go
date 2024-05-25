package server

import (
	"log"
	"net/http"

	"github.com/odin-sofware/nyusu/internal/database"
)

func (cfg *APIConfig) GetPostByUsers(w http.ResponseWriter, r *http.Request, user database.User) {
	posts, err := cfg.DB.GetPostsByUser(cfg.ctx, database.GetPostsByUserParams{
		UserID: user.ID,
		Limit:  10,
	})
	if err != nil {
		log.Print(err)
		internalServerErrorHandler(w)
		return
	}
	if len(posts) < 1 {
		notFoundHandler(w)
		return
	}
	respondWithJSON(w, http.StatusOK, posts)
}
