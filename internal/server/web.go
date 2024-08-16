package server

import (
	"html/template"
	"log"
	"net/http"

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
}

func (cfg *APIConfig) GetHome(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("html/index.html")
	if err != nil {
		panic(err)
	}
	_, err = r.Cookie(SessionCookieName)
	if err != nil {
		t.Execute(w, IndexData{
			Authenticated: false,
		})
		return
	}
	err = t.Execute(w, IndexData{
		Authenticated: true,
	})
	if err != nil {
		panic(err)
	}
}
