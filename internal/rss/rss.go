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

type AtomEntry struct {
	Text      string `xml:",chardata"`
	Title     string `xml:"title"`
	Link      struct {
		Href string `xml:"href,attr"`
	} `xml:"link"`
	Summary   string `xml:"summary"`
	Content   string `xml:"content"`
	Published string `xml:"published"`
	Updated   string `xml:"updated"`
	Author    struct {
		Name string `xml:"name"`
	} `xml:"author"`
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

type AtomFeed struct {
	XMLName  xml.Name    `xml:"feed"`
	Text     string      `xml:",chardata"`
	Title    string      `xml:"title"`
	Subtitle string      `xml:"subtitle"`
	Link     []struct {
		Href string `xml:"href,attr"`
		Rel  string `xml:"rel,attr"`
	} `xml:"link"`
	Entries []AtomEntry `xml:"entry"`
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

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Print(err)
		return Rss{}, errors.New("couldn't read the request body")
	}

	// Try RSS first
	var rssFeed *Rss
	err = xml.Unmarshal(data, &rssFeed)
	if err == nil {
		return *rssFeed, nil
	}

	// Try Atom feed
	var atomFeed *AtomFeed
	err = xml.Unmarshal(data, &atomFeed)
	if err == nil {
		// Convert Atom to RSS format
		rss := Rss{
			XMLName: xml.Name{Local: "rss"},
			Version: "2.0",
		}
		rss.Channel.Title = atomFeed.Title
		rss.Channel.Description = atomFeed.Subtitle

		// Find the alternate link
		for _, link := range atomFeed.Link {
			if link.Rel == "alternate" {
				rss.Channel.Link = link.Href
				break
			}
		}

		// Convert entries to items
		for _, entry := range atomFeed.Entries {
			item := Entry{
				Title:       entry.Title,
				Url:         entry.Link.Href,
				Description: entry.Summary,
				Content:     entry.Content,
				Published:   entry.Published,
				Author:      entry.Author.Name,
			}
			rss.Channel.Items = append(rss.Channel.Items, item)
		}

		return rss, nil
	}

	// Neither RSS nor Atom worked
	dataStr := string(data)
	if len(dataStr) > 100 {
		dataStr = dataStr[:100] + "..."
	}
	log.Printf("Feed parsing failed for URL %s. Response content: %s", url, dataStr)
	return Rss{}, errors.New("couldn't parse feed - not a valid RSS or Atom feed")
}
