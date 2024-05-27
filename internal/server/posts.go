package server

import (
	"log"
	"math"
	"net/http"
	"strconv"

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

func (cfg *APIConfig) GetPostByUsersAndFeed(w http.ResponseWriter, r *http.Request, user database.User) {
	ps, pn := GetPageSizeNumber(r)
	feedId := r.PathValue("feedId")
	id, err := strconv.ParseInt(feedId, 10, 64)
	if err != nil {
		log.Print(err)
		badRequestHandler(w)
		return
	}
	posts, err := cfg.DB.GetPostsByUserAndFeed(cfg.ctx, database.GetPostsByUserAndFeedParams{
		UserID: user.ID,
		ID:     id,
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
