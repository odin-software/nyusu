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
	Creator     string `xml:"creator"`
	Author      string `xml:"author"`
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
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return Rss{}, errors.New("couldn't create request")
	}
	// Set a proper User-Agent to avoid being blocked by servers
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; Nyusu RSS Reader/1.0)")

	resp, err := client.Do(req)
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
