package server

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/odin-sofware/nyusu/internal/database"
	"golang.org/x/crypto/bcrypt"
)

type TokenObj struct {
	Token string `json:"token"`
}

type Claims struct {
	Id int64 `json:"userId"`
	jwt.RegisteredClaims
}

func (cfg *APIConfig) LoginUser(w http.ResponseWriter, r *http.Request) {
	// w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
	var reqUser *struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	err := json.NewDecoder(r.Body).Decode(&reqUser)
	if err != nil {
		badRequestHandler(w)
		return
	}
	user, err := cfg.DB.GetUserByEmail(cfg.ctx, reqUser.Email)
	if err != nil {
		unathorizedHandler(w)
		return
	}
	b := CheckPasswordHash(reqUser.Password, user.Password)
	if !b {
		unathorizedHandler(w)
		return
	}
	token, _ := generateJWT(string(cfg.Env.SecretKey), user.ID)
	t := TokenObj{Token: token}
	respondWithJSON(w, http.StatusOK, t)
}

func (cfg *APIConfig) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var reqUser *struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	err := json.NewDecoder(r.Body).Decode(&reqUser)
	if err != nil {
		log.Print(err)
		badRequestHandler(w)
		return
	}
	hashedPassword, err := HashPassword(reqUser.Password)
	if err != nil {
		log.Print(err)
		badRequestHandler(w)
		return
	}
	user, err := cfg.DB.CreateUser(cfg.ctx, database.CreateUserParams{
		Email:    reqUser.Email,
		Password: hashedPassword,
	})
	if err != nil {
		log.Print(err)
		internalServerErrorHandler(w)
		return
	}
	respondWithJSON(w, http.StatusOK, user)
}

func (cfg *APIConfig) MiddlewareAuth(handler AuthHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: add cors from env variable.
		log.Println("HEY")
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
		header := r.Header.Get("Authorization")
		key := strings.Split(header, " ")
		if key[0] != "Bearer" || len(key) < 2 {
			unathorizedHandler(w)
			return
		}
		claims := &Claims{}
		tkn, err := jwt.ParseWithClaims(key[1], claims, func(token *jwt.Token) (interface{}, error) {
			return cfg.Env.SecretKey, nil
		})
		if err != nil {
			if err == jwt.ErrSignatureInvalid {
				unathorizedHandler(w)
				return
			}
			badRequestHandler(w)
			return
		}
		if !tkn.Valid {
			unathorizedHandler(w)
			return
		}
		user, err := cfg.DB.GetUserById(cfg.ctx, claims.Id)
		if err != nil {
			log.Print(err)
			notFoundHandler(w)
			return
		}
		handler(w, r, user)
	}
}

func generateJWT(key string, id int64) (string, error) {
	expirationTime := time.Now().Add(1 * time.Hour)
	claims := &Claims{
		Id: id,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(key))
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
