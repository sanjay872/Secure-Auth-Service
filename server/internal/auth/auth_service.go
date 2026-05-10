package auth

import (
	"context"
	"errors"
	"time"

	firebaseauth "firebase.google.com/go/v4/auth"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	DB           *pgxpool.Pool
	FirebaseAuth *firebaseauth.Client
	JWTSecret    []byte
}

func NewService(db *pgxpool.Pool, fb *firebaseauth.Client, secret []byte) *Service {
	return &Service{DB: db, FirebaseAuth: fb, JWTSecret: secret}
}

func (s *Service) CreateAccessToken(userID string, accessTTL time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(accessTTL).Unix(),
		"iat": time.Now().Unix(),
	})
	return token.SignedString(s.JWTSecret)
}

func (s *Service) CreateRefreshToken(ctx context.Context, userID string, refreshTTL time.Duration) (string, error) {
	refreshToken := uuid.NewString()
	refreshID := uuid.New()
	expiresAt := time.Now().Add(refreshTTL)

	_, err := s.DB.Exec(ctx,
		`INSERT INTO refresh_tokens (id, user_id, token, expires_at)
		 VALUES ($1, $2, $3, $4)`,
		refreshID, userID, refreshToken, expiresAt,
	)
	return refreshToken, err
}

type RefreshRecord struct {
	UserID    string
	ExpiresAt time.Time
	RevokedAt *time.Time
}

func (s *Service) GetRefreshRecord(ctx context.Context, token string) (*RefreshRecord, error) {
	var rec RefreshRecord
	err := s.DB.QueryRow(ctx,
		`SELECT user_id, expires_at, revoked_at
		   FROM refresh_tokens
		  WHERE token = $1`,
		token,
	).Scan(&rec.UserID, &rec.ExpiresAt, &rec.RevokedAt)

	if err != nil {
		return nil, err
	}
	return &rec, nil
}

func (s *Service) RevokeRefreshToken(ctx context.Context, token string) error {
	_, err := s.DB.Exec(ctx,
		`UPDATE refresh_tokens
		    SET revoked_at = $1
		  WHERE token = $2`,
		time.Now(), token,
	)
	return err
}

func (s *Service) RotateRefreshToken(ctx context.Context, oldToken string, userID string, refreshTTL time.Duration) (string, error) {
	// Revoke old
	if err := s.RevokeRefreshToken(ctx, oldToken); err != nil {
		return "", err
	}
	// Create new
	return s.CreateRefreshToken(ctx, userID, refreshTTL)
}

// Helper validation (optional but clean)
func ValidateRefreshRecord(rec *RefreshRecord) error {
	if time.Now().After(rec.ExpiresAt) {
		return errors.New("refresh token expired")
	}
	if rec.RevokedAt != nil {
		return errors.New("refresh token revoked")
	}
	return nil
}
