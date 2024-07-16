package server

import (
	"database/sql"
	"html/template"
	"log"
	"os"

	"github.com/odin-sofware/nyusu/internal/database"
	"github.com/odin-sofware/nyusu/internal/rss"
)

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func Basic() {
	posts := []database.Post{
		{
			ID:          1,
			Title:       "The best article",
			Description: sql.NullString{String: "One really cool description", Valid: true},
		},
		{
			ID:          2,
			Title:       "The worst article",
			Description: sql.NullString{String: "", Valid: false},
		},
	}
	var tmplFile = "internal/server/posts.tmpl"
	t, err := template.New("posts.tmpl").ParseFiles(tmplFile)
	checkError(err)
	err = t.Execute(os.Stdout, posts)
	checkError(err)
}

func TestRssParsing(url string) {
	r, err := rss.DataFromFeed(url)
	checkError(err)
	log.Println(r.Channel.Items[0].Creator)
}
