package server

import (
	"encoding/json"
	"math"
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

func GetPageSizeNumber(r *http.Request) (limit int64, offset int64) {
	q := r.URL.Query()
	ps := q.Get("pageSize")
	pn := q.Get("pageNumber")
	pageSize, err := strconv.ParseInt(ps, 10, 64)
	if err != nil {
		pageSize = 10
	}
	pageNumber, err := strconv.ParseInt(pn, 10, 64)
	if err != nil {
		pageNumber = 0
	}
	limit = pageSize
	offset = int64(math.Max(float64((pageNumber-1)*limit), 0.0))
	return
}
