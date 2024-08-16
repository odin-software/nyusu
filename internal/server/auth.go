package server

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/odin-sofware/nyusu/internal/database"
	"golang.org/x/crypto/bcrypt"
)

const SessionCookieName = "session_id"

type TokenObj struct {
	Token string `json:"token"`
}

type Claims struct {
	Id int64 `json:"userId"`
	jwt.RegisteredClaims
}

func (cfg *APIConfig) LoginUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	email := r.FormValue("email")
	password := r.FormValue("password")
	user, err := cfg.DB.GetUserByEmail(cfg.ctx, email)
	if err != nil {
		unathorizedHandler(w)
		return
	}
	b := CheckPasswordHash(password, user.Password)
	if !b {
		unathorizedHandler(w)
		return
	}
	cfg.SessionMng.Mutex.Lock()
	cfg.SessionMng.Sessions[email] = true
	cfg.SessionMng.Mutex.Unlock()

	// TODO: also set domain before deploying
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    email,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   60 * 60 * 24 * 3,
	})

	http.Redirect(w, r, "/", http.StatusFound)
}

func (cfg *APIConfig) LogoutUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	cookie, err := r.Cookie(SessionCookieName)
	if err == nil {
		cfg.SessionMng.Mutex.Lock()
		delete(cfg.SessionMng.Sessions, cookie.Value)
		cfg.SessionMng.Mutex.Unlock()
	}

	http.SetCookie(w, &http.Cookie{
		Name:   SessionCookieName,
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	http.Redirect(w, r, "/", http.StatusFound)
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
		cookie, err := r.Cookie(SessionCookieName)
		if err != nil {
			http.Redirect(w, r, "/", http.StatusMovedPermanently)
			return
		}
		user, err := cfg.DB.GetUserByEmail(cfg.ctx, cookie.String())
		if err != nil {
			http.Redirect(w, r, "/", http.StatusMovedPermanently)
			return
		}
		handler(w, r, user)
	}
}

func CORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Access-Control-Allow-Origin", "*")
		w.Header().Add("Access-Control-Allow-Credentials", "true")
		w.Header().Add("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		w.Header().Add("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")

		if r.Method == "OPTIONS" {
			http.Error(w, "No Content", http.StatusNoContent)
			return
		}

		next(w, r)
	}
}

func OPTIONS(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Credentials", "true")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
	w.Header().Add("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	http.Error(w, "No Content", http.StatusNoContent)
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
