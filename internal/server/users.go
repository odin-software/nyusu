package server

import (
	"net/http"

	"github.com/odin-sofware/nyusu/internal/database"
)

func (cfg *APIConfig) GetAuthUser(w http.ResponseWriter, r *http.Request, user database.User) {
	respondWithJSON(w, http.StatusOK, user)
}
