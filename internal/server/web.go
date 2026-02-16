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
		"date": func(t time.Time) string {
			return t.Format("02-01-2006")
		},
		"add": func(a, b int32) int32 {
			return a + b
		},
		"sub": func(a, b int32) int32 {
			return a - b
		},
	}
}

func TestRssParsing(url string) {
	r, err := rss.DataFromFeed(url)
	checkError(err)
	log.Println(r.Channel.Items[0].Creator)
}

type BaseData struct {
	Authenticated bool
	Branding      Branding
}

type IndexData struct {
	BaseData
	Posts      []database.GetPostsByUserWithBookmarksRow
	Pagination Pagination
}

type AuthPageData struct {
	BaseData
	Error string
}

type AddFeedData struct {
	BaseData
	Error string
}

type AllFeedsData struct {
	BaseData
	Error      string
	Feeds      []database.GetAllFeedFollowsByEmailRow
	Pagination Pagination
}

type FeedPostsData struct {
	BaseData
	Posts      []database.GetPostsByUserAndFeedWithBookmarksRow
	Pagination Pagination
}

type BookmarksData struct {
	BaseData
	Posts      []database.GetBookmarkedPostsByDateRow
	Pagination Pagination
}

func (cfg *APIConfig) getHome(w http.ResponseWriter, r *http.Request, auth AuthResult) {
	fm := getTemplateFuncMap()
	t, err := template.New("layout").Funcs(fm).ParseFiles("html/layout.html", "html/index.html")
	if err != nil {
		panic(err)
	}

	if !auth.IsAuthenticated {
		t.Execute(w, IndexData{
			BaseData: BaseData{Authenticated: false, Branding: cfg.Branding},
		})
		return
	}

	pageNumber := GetPageNumber(r)
	limit, offset := GetPageSizeNumber(r)
	posts, err := cfg.DB.GetPostsByUserWithBookmarks(cfg.ctx, database.GetPostsByUserWithBookmarksParams{
		Email:  auth.SessionData.Email,
		Limit:  limit + 1,
		Offset: offset,
	})
	if err != nil {
		log.Print(err)
		internalServerErrorHandler(w)
		return
	}

	pag := NewPagination(pageNumber, len(posts), limit)
	if len(posts) > int(limit) {
		posts = posts[:limit]
	}

	err = t.Execute(w, IndexData{
		BaseData:   BaseData{Authenticated: true, Branding: cfg.Branding},
		Posts:      posts,
		Pagination: pag,
	})
	if err != nil {
		panic(err)
	}
}

func (cfg *APIConfig) GetHome(w http.ResponseWriter, r *http.Request) {
	cfg.MiddlewareWebAuth(cfg.getHome)(w, r)
}

func (cfg *APIConfig) getAddFeed(w http.ResponseWriter, r *http.Request, auth AuthResult) {
	query := r.URL.Query()
	error := query.Get("error")
	t, err := template.ParseFiles("html/layout.html", "html/add.html")
	if err != nil {
		panic(err)
	}
	err = t.ExecuteTemplate(w, "layout", AddFeedData{
		BaseData: BaseData{Authenticated: true, Branding: cfg.Branding},
		Error:    error,
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

	pageNumber := GetPageNumber(r)
	limit, offset := GetPageSizeNumber(r)
	feeds, err := cfg.DB.GetAllFeedFollowsByEmail(cfg.ctx, database.GetAllFeedFollowsByEmailParams{
		Email:  auth.SessionData.Email,
		Limit:  limit + 1,
		Offset: offset,
	})
	if err != nil {
		log.Println(err)
		internalServerErrorHandler(w)
		return
	}

	pag := NewPagination(pageNumber, len(feeds), limit)
	if len(feeds) > int(limit) {
		feeds = feeds[:limit]
	}

	t, err := template.New("layout").Funcs(getTemplateFuncMap()).ParseFiles("html/layout.html", "html/feeds.html")
	if err != nil {
		panic(err)
	}
	err = t.ExecuteTemplate(w, "layout", AllFeedsData{
		BaseData:   BaseData{Authenticated: true, Branding: cfg.Branding},
		Error:      error,
		Feeds:      feeds,
		Pagination: pag,
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
	pageNumber := GetPageNumber(r)
	limit, offset := GetPageSizeNumber(r)
	posts, err := cfg.DB.GetPostsByUserAndFeedWithBookmarks(cfg.ctx, database.GetPostsByUserAndFeedWithBookmarksParams{
		Email:  auth.SessionData.Email,
		ID:     int64(feedId),
		Limit:  limit + 1,
		Offset: offset,
	})
	if err != nil {
		log.Println(err)
		internalServerErrorHandler(w)
		return
	}

	pag := NewPagination(pageNumber, len(posts), limit)
	if len(posts) > int(limit) {
		posts = posts[:limit]
	}

	t, err := template.New("layout.html").Funcs(fm).ParseFiles("html/layout.html", "html/feeds_posts.html")
	if err != nil {
		panic(err)
	}
	err = t.ExecuteTemplate(w, "layout", FeedPostsData{
		BaseData:   BaseData{Authenticated: true, Branding: cfg.Branding},
		Posts:      posts,
		Pagination: pag,
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

	pageNumber := GetPageNumber(r)
	limit, offset := GetPageSizeNumber(r)
	posts, err := cfg.DB.GetBookmarkedPostsByDate(cfg.ctx, database.GetBookmarkedPostsByDateParams{
		UserID: auth.SessionData.UserID2,
		Limit:  limit + 1,
		Offset: offset,
	})
	if err != nil {
		log.Println(err)
		internalServerErrorHandler(w)
		return
	}

	pag := NewPagination(pageNumber, len(posts), limit)
	if len(posts) > int(limit) {
		posts = posts[:limit]
	}

	t, err := template.New("layout.html").Funcs(fm).ParseFiles("html/layout.html", "html/bookmarks.html")
	if err != nil {
		panic(err)
	}
	err = t.ExecuteTemplate(w, "layout", BookmarksData{
		BaseData:   BaseData{Authenticated: true, Branding: cfg.Branding},
		Posts:      posts,
		Pagination: pag,
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
	err = t.ExecuteTemplate(w, "layout", BaseData{
		Authenticated: auth.IsAuthenticated,
		Branding:      cfg.Branding,
	})
	if err != nil {
		panic(err)
	}
}

func (cfg *APIConfig) GetAbout(w http.ResponseWriter, r *http.Request) {
	cfg.MiddlewareWebAuth(cfg.getAbout)(w, r)
}
