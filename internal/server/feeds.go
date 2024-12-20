package server

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"math"
	"net/http"
	"strconv"

	"github.com/odin-software/nyusu/internal/database"
	"github.com/odin-software/nyusu/internal/rss"
)

func (cfg *APIConfig) GetAllFeeds2(w http.ResponseWriter, r *http.Request) {
	ps, pn := GetPageSizeNumber(r)
	feeds, err := cfg.DB.GetAllFeeds(cfg.ctx, database.GetAllFeedsParams{
		Limit:  ps,
		Offset: int64(math.Min(float64(pn-1*ps), 0.0)),
	})
	if err != nil {
		log.Print(err)
		internalServerErrorHandler(w)
		return
	}
	if len(feeds) < 1 {
		notFoundHandler(w)
		return
	}
	respondWithJSON(w, http.StatusOK, feeds)
}

func (cfg *APIConfig) CreateFeed(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(SessionCookieName)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	url := r.FormValue("rss")
	rss, err := rss.DataFromFeed(url)
	if err != nil {
		http.Redirect(w, r, "/add?error=couldn't process url", http.StatusSeeOther)
		return
	}
	user, err := cfg.DB.GetUserByEmail(cfg.ctx, cookie.Value)
	if err != nil {
		log.Print(err)
		internalServerErrorHandler(w)
		return
	}
	feed, err := cfg.DB.CreateFeed(cfg.ctx, database.CreateFeedParams{
		Url:         url,
		Name:        rss.Channel.Title,
		Description: sql.NullString{String: rss.Channel.Description, Valid: true},
		ImageUrl:    sql.NullString{String: rss.Channel.Image.Url, Valid: true},
		ImageText:   sql.NullString{String: rss.Channel.Image.Title, Valid: true},
		Language:    sql.NullString{String: rss.Channel.Language, Valid: true},
		UserID:      user.ID,
	})
	if err != nil {
		log.Print(err)
		internalServerErrorHandler(w)
		return
	}
	_, err = cfg.DB.CreateFeedFollows(cfg.ctx, database.CreateFeedFollowsParams{
		UserID: user.ID,
		FeedID: feed.ID,
	})
	if err != nil {
		log.Print(err)
		internalServerErrorHandler(w)
		return
	}

	cfg.FetchOneFeedSync(feed.ID, feed.Url)
	http.Redirect(w, r, "/", http.StatusFound)
}

func (cfg *APIConfig) GetFeedFollowsFromUser(w http.ResponseWriter, r *http.Request, user database.User) {
	feeds, err := cfg.DB.GetFeedFollowsFromUser(cfg.ctx, user.ID)
	if err != nil {
		log.Print(err)
		internalServerErrorHandler(w)
		return
	}
	if len(feeds) < 1 {
		notFoundHandler(w)
		return
	}
	respondWithJSON(w, http.StatusOK, feeds)
}

func (cfg *APIConfig) CreateFeedFollows(w http.ResponseWriter, r *http.Request, user database.User) {
	var reqFeedFollow *struct {
		FeedId int64 `json:"feed_id,omitempty"`
	}
	err := json.NewDecoder(r.Body).Decode(&reqFeedFollow)
	if err != nil {
		badRequestHandler(w)
		return
	}
	_, err = cfg.DB.GetFeedFollows(cfg.ctx, database.GetFeedFollowsParams{
		UserID: user.ID,
		FeedID: reqFeedFollow.FeedId,
	})
	if err == nil {
		internalServerErrorHandler(w)
		return
	}
	feedFollow, err := cfg.DB.CreateFeedFollows(cfg.ctx, database.CreateFeedFollowsParams{
		UserID: user.ID,
		FeedID: reqFeedFollow.FeedId,
	})
	if err != nil {
		log.Print(err)
		internalServerErrorHandler(w)
		return
	}
	respondWithJSON(w, http.StatusCreated, feedFollow)
}

func (cfg *APIConfig) DeleteFeedFollows(w http.ResponseWriter, r *http.Request) {
	feedFollowId := r.PathValue("feedFollowId")
	id, err := strconv.Atoi(feedFollowId)
	if err != nil {
		log.Print(err)
		badRequestHandler(w)
		return
	}
	err = cfg.DB.DeleteFeedFollows(cfg.ctx, int64(id))
	if err != nil {
		log.Print(err)
		internalServerErrorHandler(w)
		return
	}
	respondOk(w)
}

func GetFeedId(r *http.Request) (int64, error) {
	q := r.URL.Query()
	fi := q.Get("feedId")
	feedId, err := strconv.ParseInt(fi, 10, 64)
	if err != nil {
		return 0, errors.New("no feedId provided")
	}
	return feedId, nil
}
