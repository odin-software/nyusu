package server

import (
	"net/http"
	"time"

	"github.com/odin-software/nyusu/internal/database"
)

type UserResponse struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (cfg *APIConfig) GetAuthUser(w http.ResponseWriter, r *http.Request, user database.User) {
	uwp := UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
	respondWithJSON(w, http.StatusOK, uwp)
}
