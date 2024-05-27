package server

import (
	"log"
	"math"
	"net/http"

	"github.com/odin-sofware/nyusu/internal/database"
)

func (cfg *APIConfig) GetPostByUsers(w http.ResponseWriter, r *http.Request, user database.User) {
	ps, pn := GetPageSizeNumber(r)
	posts, err := cfg.DB.GetPostsByUser(cfg.ctx, database.GetPostsByUserParams{
		UserID: user.ID,
		Limit:  ps,
		Offset: int64(math.Max(float64((pn-1)*ps), 0.0)),
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
