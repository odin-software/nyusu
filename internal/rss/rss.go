package rss

import (
	"encoding/xml"
	"errors"
	"io"
	"log"
	"net/http"
)

type Entry struct {
	Text        string `xml:",chardata"`
	Title       string `xml:"title"`
	Url         string `xml:"link"`
	Description string `xml:"description"`
	Published   string `xml:"pubDate"`
	Content     string `xml:"content"`
	Creator     string `xml:",cdata"`
}

type Image struct {
	Url   string `xml:"url"`
	Title string `xml:"title"`
}

type Rss struct {
	XMLName xml.Name `xml:"rss"`
	Text    string   `xml:",chardata"`
	Version string   `xml:"version,attr"`
	Channel struct {
		Text        string  `xml:",chardata"`
		Title       string  `xml:"title"`
		Link        string  `xml:"link"`
		Description string  `xml:"description"`
		Language    string  `xml:"language"`
		Image       Image   `xml:"image"`
		Items       []Entry `xml:"item"`
	} `xml:"channel"`
}

func DataFromFeed(url string) (Rss, error) {
	resp, err := http.Get(url)
	if err != nil {
		return Rss{}, errors.New("couldn't fetch the url")
	}
	defer resp.Body.Close()
	var rssFeed *Rss
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Print(err)
		return Rss{}, errors.New("couldn't read the request body")
	}
	err = xml.Unmarshal(data, &rssFeed)
	if err != nil {
		log.Print(err)
		return Rss{}, errors.New("couldn't unmarshall the data into an RSS type")
	}
	return *rssFeed, nil
}
