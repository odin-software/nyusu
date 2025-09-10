package server

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
	"time"

	"github.com/odin-software/nyusu/internal/database"
	"golang.org/x/crypto/bcrypt"
)

const SessionCookieName = "session_id"

type TokenObj struct {
	Token string `json:"token"`
}

func (cfg *APIConfig) LoginUser(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")
	user, err := cfg.DB.GetUserByEmail(cfg.ctx, email)
	if err != nil {
		http.Redirect(w, r, `/login?error=invalid credentials`, http.StatusSeeOther)
		return
	}
	b := CheckPasswordHash(password, user.Password)
	if !b {
		http.Redirect(w, r, `/login?error=invalid credentials`, http.StatusSeeOther)
		return
	}

	// Generate secure session token
	sessionToken, err := GenerateSecureToken()
	if err != nil {
		log.Print("Failed to generate session token:", err)
		http.Redirect(w, r, `/login?error=server error`, http.StatusSeeOther)
		return
	}

	// Create session in database (expires in 3 days)
	expiresAt := time.Now().Add(72 * time.Hour).Unix()
	_, err = cfg.DB.CreateSession(cfg.ctx, database.CreateSessionParams{
		Token:     sessionToken,
		UserID:    user.ID,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		log.Print("Failed to create session:", err)
		http.Redirect(w, r, `/login?error=server error`, http.StatusSeeOther)
		return
	}

	// Set secure cookie with the session token
	secure, sameSite := cfg.GetSecureCookieSettings()
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    sessionToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
		MaxAge:   60 * 60 * 24 * 3, // 3 days
	})

	http.Redirect(w, r, "/", http.StatusFound)
}

func (cfg *APIConfig) LogoutUser(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(SessionCookieName)
	if err == nil {
		err = cfg.DB.DeleteSession(cfg.ctx, cookie.Value)
		if err != nil {
			log.Print("Failed to delete session:", err)
		}
	}

	secure, sameSite := cfg.GetSecureCookieSettings()
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
		MaxAge:   -1,
	})

	http.Redirect(w, r, "/", http.StatusFound)
}

func (cfg *APIConfig) RegisterUser(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")
	confirmPassword := r.FormValue("confirmPassword")
	if password != confirmPassword {
		http.Redirect(w, r, `/register?error=passwords do not match`, http.StatusSeeOther)
		return
	}
	_, err := cfg.DB.GetUserByEmail(cfg.ctx, email)
	if err == nil {
		http.Redirect(w, r, `/register?error=user already exists`, http.StatusSeeOther)
		return
	}
	hashedPassword, err := HashPassword(password)
	if err != nil {
		log.Print(err)
		badRequestHandler(w)
		return
	}
	user, err := cfg.DB.CreateUser(cfg.ctx, database.CreateUserParams{
		Email:    email,
		Password: hashedPassword,
	})
	if err != nil {
		log.Print(err)
		internalServerErrorHandler(w)
		return
	}

	sessionToken, err := GenerateSecureToken()
	if err != nil {
		log.Print("Failed to generate session token:", err)
		http.Redirect(w, r, `/register?error=server error`, http.StatusSeeOther)
		return
	}

	// Create session in database (expires in 3 days)
	expiresAt := time.Now().Add(72 * time.Hour).Unix()
	_, err = cfg.DB.CreateSession(cfg.ctx, database.CreateSessionParams{
		Token:     sessionToken,
		UserID:    user.ID,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		log.Print("Failed to create session:", err)
		http.Redirect(w, r, `/register?error=server error`, http.StatusSeeOther)
		return
	}

	secure, sameSite := cfg.GetSecureCookieSettings()
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    sessionToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
		MaxAge:   60 * 60 * 24 * 3, // 3 days
	})

	http.Redirect(w, r, "/", http.StatusFound)
}

func (cfg *APIConfig) MiddlewareAuth(handler AuthHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(SessionCookieName)
		if err != nil {
			http.Redirect(w, r, "/", http.StatusMovedPermanently)
			return
		}

		// Get session and user from database using the secure token
		sessionData, err := cfg.DB.GetSessionByToken(cfg.ctx, cookie.Value)
		if err != nil {
			// Session not found or expired - clear the cookie and redirect
			secure, sameSite := cfg.GetSecureCookieSettings()
			http.SetCookie(w, &http.Cookie{
				Name:     SessionCookieName,
				Value:    "",
				Path:     "/",
				HttpOnly: true,
				Secure:   secure,
				SameSite: sameSite,
				MaxAge:   -1,
			})
			http.Redirect(w, r, "/", http.StatusMovedPermanently)
			return
		}

		// Convert the session data to a user object
		user := database.User{
			ID:        sessionData.ID_2,
			Name:      sessionData.Name,
			Email:     sessionData.Email,
			Password:  sessionData.Password,
			CreatedAt: sessionData.CreatedAt_2,
			UpdatedAt: sessionData.UpdatedAt,
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

func GenerateSecureToken() (string, error) {
	bytes := make([]byte, 32) // 32 bytes = 256 bits of entropy
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func (cfg *APIConfig) GetSecureCookieSettings() (secure bool, sameSite http.SameSite) {
	if cfg.Env.Environment == "production" {
		return true, http.SameSiteStrictMode
	}
	return false, http.SameSiteLaxMode
}

func (cfg *APIConfig) CleanupExpiredSessions() {
	err := cfg.DB.DeleteExpiredSessions(cfg.ctx)
	if err != nil {
		log.Print("Failed to cleanup expired sessions:", err)
	}
}
