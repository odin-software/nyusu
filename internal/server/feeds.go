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
	url := SanitizeInput(r.FormValue("rss"))
	if url == "" {
		http.Redirect(w, r, "/add?error=RSS URL is required", http.StatusSeeOther)
		return
	}

	rssData, err := rss.DataFromFeed(url)
	if err != nil {
		http.Redirect(w, r, "/add?error=couldn't process url", http.StatusSeeOther)
		return
	}
	sessionData, err := cfg.DB.GetSessionByToken(cfg.ctx, cookie.Value)
	if err != nil {
		log.Print("Invalid session:", err)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	user := database.User{
		ID:        sessionData.ID_2,
	}

	existingFeed, err := cfg.DB.GetFeedByUrl(cfg.ctx, url)
	var feed database.Feed

	if err != nil {
		// Feed doesn't exist, create a new one
		feed, err = cfg.DB.CreateFeed(cfg.ctx, database.CreateFeedParams{
			Url:         url,
			Name:        rssData.Channel.Title,
			Link:        sql.NullString{String: rssData.Channel.Link, Valid: rssData.Channel.Link != ""},
			Description: sql.NullString{String: rssData.Channel.Description, Valid: true},
			ImageUrl:    sql.NullString{String: rssData.Channel.Image.Url, Valid: true},
			ImageText:   sql.NullString{String: rssData.Channel.Image.Title, Valid: true},
			Language:    sql.NullString{String: rssData.Channel.Language, Valid: true},
			UserID:      user.ID,
		})
		if err != nil {
			log.Print(err)
			internalServerErrorHandler(w)
			return
		}
	} else {
		feed = existingFeed
	}

	_, err = cfg.DB.GetFeedFollows(cfg.ctx, database.GetFeedFollowsParams{
		UserID: user.ID,
		FeedID: feed.ID,
	})
	if err == nil {
		http.Redirect(w, r, "/add?error=you're already following this feed", http.StatusSeeOther)
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

func (cfg *APIConfig) UnsubscribeFeed(w http.ResponseWriter, r *http.Request) {
	feedFollowId := r.PathValue("feedFollowId")
	id, err := strconv.Atoi(feedFollowId)
	if err != nil {
		log.Print(err)
		http.Redirect(w, r, "/feeds?error=invalid feed follow ID", http.StatusSeeOther)
		return
	}
	err = cfg.DB.DeleteFeedFollows(cfg.ctx, int64(id))
	if err != nil {
		log.Print(err)
		http.Redirect(w, r, "/feeds?error=failed to unsubscribe from feed", http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/feeds", http.StatusSeeOther)
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
