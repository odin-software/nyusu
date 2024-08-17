package server

import (
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/odin-sofware/nyusu/internal/database"
	"github.com/odin-sofware/nyusu/internal/rss"
)

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func TestRssParsing(url string) {
	r, err := rss.DataFromFeed(url)
	checkError(err)
	log.Println(r.Channel.Items[0].Creator)
}

type IndexData struct {
	Authenticated bool
	Posts         []database.GetPostsByUserRow
}

type AuthPageData struct {
	Register bool
	Error    string
}

type AddFeedData struct {
	Authenticated bool
	Error         string
}

func (cfg *APIConfig) GetHome(w http.ResponseWriter, r *http.Request) {
	fm := template.FuncMap{
		"date": func(i int64) string {
			t := time.Unix(i, 0)
			return t.Format("02-01-2006")
		},
	}
	t, err := template.New("index.html").Funcs(fm).ParseFiles("html/index.html")
	if err != nil {
		panic(err)
	}
	cookie, err := r.Cookie(SessionCookieName)
	if err != nil {
		t.Execute(w, IndexData{
			Authenticated: false,
		})
		return
	}
	limit, offset := GetPageSizeNumber(r)
	posts, err := cfg.DB.GetPostsByUser(cfg.ctx, database.GetPostsByUserParams{
		Email:  cookie.Value,
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
			Posts:         []database.GetPostsByUserRow{},
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

func (cfg *APIConfig) GetLogin(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	error := query.Get("error")
	_, err := r.Cookie(SessionCookieName)
	if err == nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	t, err := template.ParseFiles("html/auth.html")
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

func (cfg *APIConfig) GetRegister(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	error := query.Get("error")
	_, err := r.Cookie(SessionCookieName)
	if err == nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	t, err := template.ParseFiles("html/auth.html")
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

func (cfg *APIConfig) GetAddFeed(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	error := query.Get("error")
	_, err := r.Cookie(SessionCookieName)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
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
