package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

type Handler struct {
	Service *Service
}

func NewHandler(s *Service) *Handler {
	return &Handler{Service: s}
}

type ExchangeRequest struct {
	IDToken string `json:"id_token"`
}

const (
	AccessTTLDev  = 15 * time.Second
	AccessTTLProd = 15 * time.Minute

	RefreshTTLProd = 7 * 24 * time.Hour
)

func setRefreshCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    token,
		HttpOnly: true,
		Secure:   false,                // true in production (HTTPS)
		SameSite: http.SameSiteLaxMode, // Lax for localhost dev
		Path:     "/",
		MaxAge:   int(RefreshTTLProd.Seconds()),
	})
}

func clearRefreshCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
		MaxAge:   -1,
	})
}

func (h *Handler) Exchange(w http.ResponseWriter, r *http.Request) {
	var req ExchangeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ctx := context.Background()

	// Verify Firebase ID token
	token, err := h.Service.FirebaseAuth.VerifyIDToken(ctx, req.IDToken)
	if err != nil {
		http.Error(w, "Invalid Firebase token", http.StatusUnauthorized)
		return
	}

	userID := token.UID

	// Issue tokens
	accessToken, err := h.Service.CreateAccessToken(userID, AccessTTLDev) // change to AccessTTLProd later
	if err != nil {
		http.Error(w, "Failed to create access token", http.StatusInternalServerError)
		return
	}

	refreshToken, err := h.Service.CreateRefreshToken(ctx, userID, RefreshTTLProd)
	if err != nil {
		http.Error(w, "Failed to store refresh token", http.StatusInternalServerError)
		return
	}

	setRefreshCookie(w, refreshToken)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"access_token": accessToken,
	})
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	c, err := r.Cookie("refresh_token")
	if err != nil {
		http.Error(w, "Missing refresh token", http.StatusUnauthorized)
		return
	}

	oldToken := c.Value

	rec, err := h.Service.GetRefreshRecord(ctx, oldToken)
	if err != nil {
		http.Error(w, "Invalid refresh token", http.StatusUnauthorized)
		return
	}

	if err := ValidateRefreshRecord(rec); err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// Rotate refresh token
	newRefresh, err := h.Service.RotateRefreshToken(ctx, oldToken, rec.UserID, RefreshTTLProd)
	if err != nil {
		http.Error(w, "Failed to rotate refresh token", http.StatusInternalServerError)
		return
	}

	// New access token
	newAccess, err := h.Service.CreateAccessToken(rec.UserID, AccessTTLDev) // change to AccessTTLProd later
	if err != nil {
		http.Error(w, "Failed to create access token", http.StatusInternalServerError)
		return
	}

	setRefreshCookie(w, newRefresh)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"access_token": newAccess,
	})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	c, err := r.Cookie("refresh_token")
	if err != nil {
		http.Error(w, "No refresh token", http.StatusUnauthorized)
		return
	}

	_ = h.Service.RevokeRefreshToken(ctx, c.Value) // even if fails, still clear cookie

	clearRefreshCookie(w)
	w.Write([]byte("Logged out successfully"))
}

func (h *Handler) Profile(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"user_id": userID,
		"message": "Protected profile data 🔐",
	})
}
