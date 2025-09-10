package server

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/odin-software/nyusu/internal/database"
	"golang.org/x/crypto/bcrypt"
)

const SessionCookieName = "session_id"

type TokenObj struct {
	Token string `json:"token"`
}

func (cfg *APIConfig) LoginUser(w http.ResponseWriter, r *http.Request) {
	email := SanitizeInput(r.FormValue("email"))
	password := r.FormValue("password")

	// Basic validation
	if email == "" || password == "" {
		http.Redirect(w, r, `/login?error=email and password are required`, http.StatusSeeOther)
		return
	}

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
	email := SanitizeInput(r.FormValue("email"))
	password := r.FormValue("password")
	confirmPassword := r.FormValue("confirmPassword")

	if !ValidateEmail(email) {
		http.Redirect(w, r, `/register?error=invalid email format`, http.StatusSeeOther)
		return
	}

	if valid, errMsg := ValidatePassword(password); !valid {
		http.Redirect(w, r, `/register?error=`+errMsg, http.StatusSeeOther)
		return
	}

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

// CORS returns a CORS middleware configured for the current environment
func (cfg *APIConfig) CORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		allowOrigin, allowCredentials := cfg.GetCORSSettings()

		w.Header().Set("Access-Control-Allow-Origin", allowOrigin)

		if allowCredentials {
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")

		if r.Method == "OPTIONS" {
			http.Error(w, "No Content", http.StatusNoContent)
			return
		}

		next(w, r)
	}
}

func (cfg *APIConfig) OPTIONS(w http.ResponseWriter, r *http.Request) {
	allowOrigin, allowCredentials := cfg.GetCORSSettings()

	w.Header().Set("Access-Control-Allow-Origin", allowOrigin)

	if allowCredentials {
		w.Header().Set("Access-Control-Allow-Credentials", "true")
	}

	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")

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

// ValidateEmail checks if email format is valid
func ValidateEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// ValidatePassword checks password strength requirements
func ValidatePassword(password string) (bool, string) {
	if len(password) < 8 {
		return false, "password must be at least 8 characters long"
	}

	if len(password) > 128 {
		return false, "password must be less than 128 characters long"
	}

	var hasUpper, hasLower, hasNumber, hasSpecial bool

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return false, "password must contain at least one uppercase letter"
	}
	if !hasLower {
		return false, "password must contain at least one lowercase letter"
	}
	if !hasNumber {
		return false, "password must contain at least one number"
	}
	if !hasSpecial {
		return false, "password must contain at least one special character"
	}

	return true, ""
}

func SanitizeInput(input string) string {
	input = strings.TrimSpace(input)

	input = strings.ReplaceAll(input, "\x00", "")

	var result strings.Builder
	for _, char := range input {
		if unicode.IsControl(char) && char != '\n' && char != '\t' {
			continue
		}
		result.WriteRune(char)
	}

	return result.String()
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

// GetCORSSettings returns appropriate CORS settings based on environment
func (cfg *APIConfig) GetCORSSettings() (allowOrigin string, allowCredentials bool) {
	if cfg.Env.Environment == "production" {
		return cfg.Env.ProductionURL, true
	}
	return "http://localhost:8888", true
}

func (cfg *APIConfig) CleanupExpiredSessions() {
	err := cfg.DB.DeleteExpiredSessions(cfg.ctx)
	if err != nil {
		log.Print("Failed to cleanup expired sessions:", err)
	}
}
