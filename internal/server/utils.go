package server

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"errors"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strconv"
)

func internalServerErrorHandler(w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("500 Internal Server Error"))
}

func badRequestHandler(w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte("400 Malformed Request"))
}

func notFoundHandler(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("404 Not Found"))
}

func unathorizedHandler(w http.ResponseWriter) {
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte("401 Unauthorized"))
}

func respondOk(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
}

func respondWithJSON(w http.ResponseWriter, status int, payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		internalServerErrorHandler(w)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(data)
}

type Rss struct {
	XMLName xml.Name `xml:"rss"`
	Text    string   `xml:",chardata"`
	Version string   `xml:"version,attr"`
	Channel struct {
		Text        string `xml:",chardata"`
		Title       string `xml:"title"`
		Link        string `xml:"link"`
		Description string `xml:"description"`
		Items       []struct {
			Text        string `xml:",chardata"`
			Title       string `xml:"title"`
			Description string `xml:"description"`
			Published   string `xml:"pubDate"`
			Url         string `xml:"link"`
		} `xml:"item"`
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
	}
	err = xml.Unmarshal(data, &rssFeed)
	if err != nil {
		log.Print(err)
	}
	return *rssFeed, nil
}

func GetNewHash() string {
	r := strconv.FormatFloat(rand.Float64(), 'f', -1, 64)
	h := sha256.New()
	h.Write([]byte(r))

	return hex.EncodeToString(h.Sum((nil)))
}
