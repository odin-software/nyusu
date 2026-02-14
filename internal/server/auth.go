package server

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
	"strings"
	"time"
	"unicode"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/odin-software/nyusu/internal/database"
)

const SessionCookieName = "session_id"
const OIDCStateCookieName = "oidc_state"

// LoginRedirect initiates the OIDC login flow by redirecting to the identity provider.
func (cfg *APIConfig) LoginRedirect(w http.ResponseWriter, r *http.Request) {
	// Check if already authenticated
	cookie, err := r.Cookie(SessionCookieName)
	if err == nil {
		_, err := cfg.DB.GetSessionByToken(cfg.ctx, cookie.Value)
		if err == nil {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
	}

	state, err := GenerateSecureToken()
	if err != nil {
		log.Print("Failed to generate OIDC state:", err)
		internalServerErrorHandler(w)
		return
	}

	secure, sameSite := cfg.GetSecureCookieSettings()
	http.SetCookie(w, &http.Cookie{
		Name:     OIDCStateCookieName,
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
		MaxAge:   300, // 5 minutes
	})

	http.Redirect(w, r, cfg.OAuth2Config.AuthCodeURL(state), http.StatusFound)
}

// OIDCCallback handles the callback from the identity provider after authentication.
func (cfg *APIConfig) OIDCCallback(w http.ResponseWriter, r *http.Request) {
	// Verify state parameter
	stateCookie, err := r.Cookie(OIDCStateCookieName)
	if err != nil {
		log.Print("Missing OIDC state cookie")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if r.URL.Query().Get("state") != stateCookie.Value {
		log.Print("OIDC state mismatch")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Clear state cookie
	secure, sameSite := cfg.GetSecureCookieSettings()
	http.SetCookie(w, &http.Cookie{
		Name:     OIDCStateCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
		MaxAge:   -1,
	})

	// Exchange code for tokens
	oauth2Token, err := cfg.OAuth2Config.Exchange(cfg.ctx, r.URL.Query().Get("code"))
	if err != nil {
		log.Print("Failed to exchange OIDC code:", err)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Extract and verify ID token
	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		log.Print("No id_token in OIDC response")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	verifier := cfg.OIDCProvider.Verifier(&oidc.Config{ClientID: cfg.Env.OIDCClientID})
	idToken, err := verifier.Verify(cfg.ctx, rawIDToken)
	if err != nil {
		log.Print("Failed to verify OIDC ID token:", err)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Extract claims
	var claims struct {
		Sub   string `json:"sub"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := idToken.Claims(&claims); err != nil {
		log.Print("Failed to parse OIDC claims:", err)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Get or create user by OIDC subject
	user, err := cfg.DB.GetOrCreateUserBySub(cfg.ctx, database.GetOrCreateUserBySubParams{
		Name:  claims.Name,
		Email: claims.Email,
		Sub:   claims.Sub,
	})
	if err != nil {
		log.Print("Failed to get/create user:", err)
		internalServerErrorHandler(w)
		return
	}

	// Generate session token
	sessionToken, err := GenerateSecureToken()
	if err != nil {
		log.Print("Failed to generate session token:", err)
		internalServerErrorHandler(w)
		return
	}

	// Create session in database (expires in 3 days)
	expiresAt := time.Now().Add(72 * time.Hour)
	_, err = cfg.DB.CreateSession(cfg.ctx, database.CreateSessionParams{
		Token:     sessionToken,
		UserID:    user.ID,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		log.Print("Failed to create session:", err)
		internalServerErrorHandler(w)
		return
	}

	// Set session cookie
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
			ID:        sessionData.UserID2,
			Name:      sessionData.Name,
			Email:     sessionData.Email,
			Sub:       sessionData.Sub,
			CreatedAt: sessionData.UserCreatedAt,
			UpdatedAt: sessionData.UserUpdatedAt,
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
