package server

import (
	"database/sql"
	"html/template"
	"log"
	"os"

	"github.com/odin-sofware/nyusu/internal/database"
)

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
	check := func(err error) {
		if err != nil {
			log.Fatal(err)
		}
	}
	t, err := template.New("posts.tmpl").ParseFiles(tmplFile)
	check(err)
	err = t.Execute(os.Stdout, posts)
	check(err)
}
