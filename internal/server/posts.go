package server

import (
	"log"
	"net/http"
	"strconv"

	"github.com/odin-software/nyusu/internal/database"
)

func (cfg *APIConfig) GetBookmarkedPosts(w http.ResponseWriter, r *http.Request, user database.User) {
	limit, offset := GetPageSizeNumber(r)
	q := r.URL.Query()
	createdOrPublished := q.Get("order")
	if createdOrPublished == "created" {
		posts, err := cfg.DB.GetBookmarkedPostsByDate(cfg.ctx, database.GetBookmarkedPostsByDateParams{
			UserID: user.ID,
			Limit:  limit,
			Offset: offset,
		})
		if err != nil {
			log.Print(err)
			internalServerErrorHandler(w)
			return
		}
		if len(posts) < 1 {
			respondWithJSON(w, http.StatusOK, []int{})
			return
		}
		respondWithJSON(w, http.StatusOK, posts)
	} else {
		posts, err := cfg.DB.GetBookmarkedPostsByPublished(cfg.ctx, database.GetBookmarkedPostsByPublishedParams{
			UserID: user.ID,
			Limit:  limit,
			Offset: offset,
		})
		if err != nil {
			log.Print(err)
			internalServerErrorHandler(w)
			return
		}
		if len(posts) < 1 {
			respondWithJSON(w, http.StatusOK, []int{})
			return
		}
		respondWithJSON(w, http.StatusOK, posts)
	}
}

func (cfg *APIConfig) BookmarkPost(w http.ResponseWriter, r *http.Request, user database.User) {
	postId := r.PathValue("postId")
	id, err := strconv.ParseInt(postId, 10, 64)
	if err != nil {
		log.Print(err)
		badRequestHandler(w)
		return
	}
	err = cfg.DB.BookmarkPost(cfg.ctx, database.BookmarkPostParams{
		UserID: user.ID,
		PostID: id,
	})
	if err != nil {
		log.Println(err)
		internalServerErrorHandler(w)
		return
	}
	respondOk(w)
}

func (cfg *APIConfig) UnbookmarkPost(w http.ResponseWriter, r *http.Request, user database.User) {
	postId := r.PathValue("postId")
	id, err := strconv.ParseInt(postId, 10, 64)
	if err != nil {
		log.Print(err)
		badRequestHandler(w)
		return
	}
	err = cfg.DB.UnbookmarkPost(cfg.ctx, database.UnbookmarkPostParams{
		UserID: user.ID,
		PostID: id,
	})
	if err != nil {
		log.Println(err)
		internalServerErrorHandler(w)
		return
	}
	respondOk(w)
}
