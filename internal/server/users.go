package server

import (
	"net/http"

	"github.com/odin-software/nyusu/internal/database"
)

type UserWithoutPassword struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
}

func (cfg *APIConfig) GetAuthUser(w http.ResponseWriter, r *http.Request, user database.User) {
	uwp := UserWithoutPassword{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
	respondWithJSON(w, http.StatusOK, uwp)
}
