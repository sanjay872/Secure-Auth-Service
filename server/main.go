package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	firebase "firebase.google.com/go/v4"
	firebaseauth "firebase.google.com/go/v4/auth"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/api/option"
)

var ctx = context.Background()
var jwtSecret = []byte("super-secret-key") // move to env in production
var db *pgxpool.Pool
var firebaseAuth *firebaseauth.Client

type ExchangeRequest struct {
	IDToken string `json:"id_token"`
}

type ExchangeResponse struct {
	AccessToken string `json:"access_token"`
}

func main() {

	// ----------------------------
	// PostgreSQL Connection
	// ----------------------------
	databaseURL := "postgres://postgres:admin@localhost:5433/secure_auth"

	var err error
	db, err = pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// ----------------------------
	// Firebase Admin Setup
	// ----------------------------
	opt := option.WithCredentialsFile("firebase-service-account.json")
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		log.Fatal("Failed to initialize Firebase:", err)
	}

	firebaseAuth, err = app.Auth(ctx)
	if err != nil {
		log.Fatal("Failed to initialize Firebase Auth:", err)
	}

	// ----------------------------
	// Router Setup
	// ----------------------------
	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true, // required for cookies
	}))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Server running ✅"))
	})

	authLimiter := rateLimitMiddleware(5, time.Minute)

	r.With(authLimiter).Post("/auth/exchange", exchangeHandler)
	r.With(authLimiter).Post("/auth/refresh", refreshHandler)
	r.With(authLimiter).Post("/auth/logout", logoutHandler)

	r.Group(func(r chi.Router) {
		r.Use(jwtMiddleware)
		r.Get("/profile", profileHandler)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("Server started on port", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

// ------------------------------
// Simple In-Memory Rate Limiter
// ------------------------------

var requestCounts = make(map[string]int)
var lastReset = time.Now()

func rateLimitMiddleware(limit int, window time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			// Reset counts if window expired
			if time.Since(lastReset) > window {
				requestCounts = make(map[string]int)
				lastReset = time.Now()
			}

			ip := r.RemoteAddr

			requestCounts[ip]++

			if requestCounts[ip] > limit {
				http.Error(w, "Too many requests", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// ----------------------------------------------------
// Exchange Firebase ID Token for App Tokens
// ----------------------------------------------------

func exchangeHandler(w http.ResponseWriter, r *http.Request) {

	var req ExchangeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 1️⃣ Verify Firebase ID Token
	token, err := firebaseAuth.VerifyIDToken(ctx, req.IDToken)
	if err != nil {
		http.Error(w, "Invalid Firebase token", http.StatusUnauthorized)
		return
	}

	userID := token.UID
	email, _ := token.Claims["email"].(string)

	// 2️⃣ Generate Access Token (15 min)
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":   userID,
		"email": email,
		"exp":   time.Now().Add(15 * time.Second).Unix(), // token expriration time
		"iat":   time.Now().Unix(),
	})

	accessTokenString, err := accessToken.SignedString(jwtSecret)
	if err != nil {
		http.Error(w, "Failed to sign access token", http.StatusInternalServerError)
		return
	}

	// 3️⃣ Generate Refresh Token (opaque UUID)
	refreshToken := uuid.NewString()
	refreshID := uuid.New()
	expiresAt := time.Now().Add(7 * 24 * time.Hour) // refresh token expiration time

	_, err = db.Exec(ctx,
		`INSERT INTO refresh_tokens (id, user_id, token, expires_at)
		 VALUES ($1, $2, $3, $4)`,
		refreshID, userID, refreshToken, expiresAt,
	)
	if err != nil {
		http.Error(w, "Failed to store refresh token", http.StatusInternalServerError)
		return
	}

	// 4️⃣ Set Refresh Token as httpOnly Cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		HttpOnly: true,
		Secure:   false, // change to true in production (HTTPS)
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
		MaxAge:   7 * 24 * 60 * 60,
	})

	// 5️⃣ Return Access Token
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ExchangeResponse{
		AccessToken: accessTokenString,
	})
}

func jwtMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
			return
		}

		// Expecting: "Bearer <token>"
		const prefix = "Bearer "
		if len(authHeader) <= len(prefix) || authHeader[:len(prefix)] != prefix {
			http.Error(w, "Invalid Authorization format", http.StatusUnauthorized)
			return
		}

		tokenString := authHeader[len(prefix):]

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "Invalid token claims", http.StatusUnauthorized)
			return
		}

		userID, ok := claims["sub"].(string)
		if !ok {
			http.Error(w, "Invalid user ID", http.StatusUnauthorized)
			return
		}

		// Add userID to request context
		ctx := context.WithValue(r.Context(), "userID", userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func profileHandler(w http.ResponseWriter, r *http.Request) {

	userID := r.Context().Value("userID").(string)

	response := map[string]string{
		"user_id": userID,
		"message": "This is protected profile data 🔐",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func refreshHandler(w http.ResponseWriter, r *http.Request) {

	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		http.Error(w, "Missing refresh token", http.StatusUnauthorized)
		return
	}

	refreshToken := cookie.Value

	var userID string
	var expiresAt time.Time
	var revokedAt *time.Time

	err = db.QueryRow(ctx,
		`SELECT user_id, expires_at, revoked_at
		 FROM refresh_tokens
		 WHERE token = $1`,
		refreshToken,
	).Scan(&userID, &expiresAt, &revokedAt)

	if err != nil {
		http.Error(w, "Invalid refresh token", http.StatusUnauthorized)
		return
	}

	// Check expiry
	if time.Now().After(expiresAt) {
		http.Error(w, "Refresh token expired", http.StatusUnauthorized)
		return
	}

	// Check revocation
	if revokedAt != nil {
		http.Error(w, "Refresh token revoked", http.StatusUnauthorized)
		return
	}

	// 🔁 ROTATION: revoke old token
	now := time.Now()
	_, err = db.Exec(ctx,
		`UPDATE refresh_tokens
		 SET revoked_at = $1
		 WHERE token = $2`,
		now, refreshToken,
	)
	if err != nil {
		http.Error(w, "Failed to revoke old token", http.StatusInternalServerError)
		return
	}

	// Generate new refresh token
	newRefreshToken := uuid.NewString()
	newRefreshID := uuid.New()
	newExpiresAt := time.Now().Add(7 * 24 * time.Hour)

	_, err = db.Exec(ctx,
		`INSERT INTO refresh_tokens (id, user_id, token, expires_at)
		 VALUES ($1, $2, $3, $4)`,
		newRefreshID, userID, newRefreshToken, newExpiresAt,
	)
	if err != nil {
		http.Error(w, "Failed to create new refresh token", http.StatusInternalServerError)
		return
	}

	// Generate new access token
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(15 * time.Minute).Unix(),
		"iat": time.Now().Unix(),
	})

	accessTokenString, err := accessToken.SignedString(jwtSecret)
	if err != nil {
		http.Error(w, "Failed to sign access token", http.StatusInternalServerError)
		return
	}

	// Set new refresh cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    newRefreshToken,
		HttpOnly: true,
		Secure:   false, // true in production
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
		MaxAge:   7 * 24 * 60 * 60,
	})

	json.NewEncoder(w).Encode(map[string]string{
		"access_token": accessTokenString,
	})
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {

	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		http.Error(w, "No refresh token", http.StatusUnauthorized)
		return
	}

	refreshToken := cookie.Value

	// Revoke token
	_, err = db.Exec(ctx,
		`UPDATE refresh_tokens
		 SET revoked_at = $1
		 WHERE token = $2`,
		time.Now(), refreshToken,
	)
	if err != nil {
		http.Error(w, "Failed to revoke token", http.StatusInternalServerError)
		return
	}

	// Clear cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		MaxAge:   -1,
	})

	w.Write([]byte("Logged out successfully"))
}
