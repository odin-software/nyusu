package server

import (
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/odin-software/nyusu/internal/database"
	"github.com/odin-software/nyusu/internal/rss"
)

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func getTemplateFuncMap() template.FuncMap {
	return template.FuncMap{
		"date": func(i int64) string {
			t := time.Unix(i, 0)
			return t.Format("02-01-2006")
		},
	}
}

func TestRssParsing(url string) {
	r, err := rss.DataFromFeed(url)
	checkError(err)
	log.Println(r.Channel.Items[0].Creator)
}

type IndexData struct {
	Authenticated bool
	Posts         []database.GetPostsByUserWithBookmarksRow
}

type AuthPageData struct {
	Authenticated bool
	Register      bool
	Error         string
}

type AddFeedData struct {
	Authenticated bool
	Error         string
}

type AllFeedsData struct {
	Authenticated bool
	Error         string
	Feeds         []database.GetAllFeedFollowsByEmailRow
}

type FeedPostsData struct {
	Authenticated bool
	Posts         []database.GetPostsByUserAndFeedWithBookmarksRow
}

type BookmarksData struct {
	Authenticated bool
	Posts         []database.GetBookmarkedPostsByDateRow
}

func (cfg *APIConfig) getHome(w http.ResponseWriter, r *http.Request, auth AuthResult) {
	fm := getTemplateFuncMap()
	t, err := template.New("layout").Funcs(fm).ParseFiles("html/layout.html", "html/index.html")
	if err != nil {
		panic(err)
	}

	if !auth.IsAuthenticated {
		t.Execute(w, IndexData{
			Authenticated: false,
		})
		return
	}

	limit, offset := GetPageSizeNumber(r)
	posts, err := cfg.DB.GetPostsByUserWithBookmarks(cfg.ctx, database.GetPostsByUserWithBookmarksParams{
		Email:  auth.SessionData.Email,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		log.Print(err)
		internalServerErrorHandler(w)
		return
	}
	if len(posts) < 1 {
		t.Execute(w, IndexData{
			Authenticated: true,
			Posts:         []database.GetPostsByUserWithBookmarksRow{},
		})
		return
	}
	err = t.Execute(w, IndexData{
		Authenticated: true,
		Posts:         posts,
	})
	if err != nil {
		panic(err)
	}
}

func (cfg *APIConfig) GetHome(w http.ResponseWriter, r *http.Request) {
	cfg.MiddlewareWebAuth(cfg.getHome)(w, r)
}

func (cfg *APIConfig) getLogin(w http.ResponseWriter, r *http.Request, auth AuthResult) {
	query := r.URL.Query()
	error := query.Get("error")
	t, err := template.New("layout").ParseFiles("html/layout.html", "html/auth.html")
	if err != nil {
		panic(err)
	}
	err = t.Execute(w, AuthPageData{
		Register: false,
		Error:    error,
	})
	if err != nil {
		panic(err)
	}
}

func (cfg *APIConfig) GetLogin(w http.ResponseWriter, r *http.Request) {
	cfg.RedirectIfAuth(cfg.getLogin)(w, r)
}

func (cfg *APIConfig) getRegister(w http.ResponseWriter, r *http.Request, auth AuthResult) {
	query := r.URL.Query()
	error := query.Get("error")
	t, err := template.New("layout").ParseFiles("html/layout.html", "html/auth.html")
	if err != nil {
		panic(err)
	}
	err = t.Execute(w, AuthPageData{
		Register: true,
		Error:    error,
	})
	if err != nil {
		panic(err)
	}
}

func (cfg *APIConfig) GetRegister(w http.ResponseWriter, r *http.Request) {
	cfg.RedirectIfAuth(cfg.getRegister)(w, r)
}

func (cfg *APIConfig) getAddFeed(w http.ResponseWriter, r *http.Request, auth AuthResult) {
	query := r.URL.Query()
	error := query.Get("error")
	t, err := template.ParseFiles("html/layout.html", "html/add.html")
	if err != nil {
		panic(err)
	}
	err = t.ExecuteTemplate(w, "layout", AddFeedData{
		Authenticated: true,
		Error:         error,
	})
	if err != nil {
		panic(err)
	}
}

func (cfg *APIConfig) GetAddFeed(w http.ResponseWriter, r *http.Request) {
	cfg.RequireAuth(cfg.getAddFeed)(w, r)
}

func (cfg *APIConfig) getAllFeeds(w http.ResponseWriter, r *http.Request, auth AuthResult) {
	query := r.URL.Query()
	error := query.Get("error")

	limit, offset := GetPageSizeNumber(r)
	feeds, err := cfg.DB.GetAllFeedFollowsByEmail(cfg.ctx, database.GetAllFeedFollowsByEmailParams{
		Email:  auth.SessionData.Email,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		log.Println(err)
		internalServerErrorHandler(w)
		return
	}
	t, err := template.ParseFiles("html/layout.html", "html/feeds.html")
	if err != nil {
		panic(err)
	}
	err = t.ExecuteTemplate(w, "layout", AllFeedsData{
		Authenticated: true,
		Error:         error,
		Feeds:         feeds,
	})
	if err != nil {
		panic(err)
	}
}

func (cfg *APIConfig) GetAllFeeds(w http.ResponseWriter, r *http.Request) {
	cfg.RequireAuth(cfg.getAllFeeds)(w, r)
}

func (cfg *APIConfig) getFeedPosts(w http.ResponseWriter, r *http.Request, auth AuthResult) {
	fm := getTemplateFuncMap()
	feed := r.PathValue("feedId")

	feedId, err := strconv.Atoi(feed)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusMovedPermanently)
		return
	}
	limit, offset := GetPageSizeNumber(r)
	posts, err := cfg.DB.GetPostsByUserAndFeedWithBookmarks(cfg.ctx, database.GetPostsByUserAndFeedWithBookmarksParams{
		Email:  auth.SessionData.Email,
		ID:     int64(feedId),
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		log.Println(err)
		internalServerErrorHandler(w)
		return
	}
	t, err := template.New("layout.html").Funcs(fm).ParseFiles("html/layout.html", "html/feeds_posts.html")
	if err != nil {
		panic(err)
	}
	err = t.ExecuteTemplate(w, "layout", FeedPostsData{
		Authenticated: true,
		Posts:         posts,
	})
	if err != nil {
		panic(err)
	}
}

func (cfg *APIConfig) GetFeedPosts(w http.ResponseWriter, r *http.Request) {
	cfg.RequireAuth(cfg.getFeedPosts)(w, r)
}

func (cfg *APIConfig) getBookmarks(w http.ResponseWriter, r *http.Request, auth AuthResult) {
	fm := getTemplateFuncMap()

	limit, offset := GetPageSizeNumber(r)
	posts, err := cfg.DB.GetBookmarkedPostsByDate(cfg.ctx, database.GetBookmarkedPostsByDateParams{
		UserID: auth.SessionData.ID_2,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		log.Println(err)
		internalServerErrorHandler(w)
		return
	}

	t, err := template.New("layout.html").Funcs(fm).ParseFiles("html/layout.html", "html/bookmarks.html")
	if err != nil {
		panic(err)
	}
	err = t.ExecuteTemplate(w, "layout", BookmarksData{
		Authenticated: true,
		Posts:         posts,
	})
	if err != nil {
		panic(err)
	}
}

func (cfg *APIConfig) GetBookmarks(w http.ResponseWriter, r *http.Request) {
	cfg.RequireAuth(cfg.getBookmarks)(w, r)
}

func (cfg *APIConfig) getAbout(w http.ResponseWriter, r *http.Request, auth AuthResult) {
	t, err := template.ParseFiles("html/layout.html", "html/about.html")
	if err != nil {
		panic(err)
	}
	err = t.ExecuteTemplate(w, "layout", struct {
		Authenticated bool
	}{
		Authenticated: auth.IsAuthenticated,
	})
	if err != nil {
		panic(err)
	}
}

func (cfg *APIConfig) GetAbout(w http.ResponseWriter, r *http.Request) {
	cfg.MiddlewareWebAuth(cfg.getAbout)(w, r)
}
