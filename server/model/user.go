package model

import (
	"database/sql"
	"time"
)

type User struct {
	Id              int64
	FirstName       string
	LastName        string
	ProfilePic      sql.NullString
	Email           string
	PasswordHash    string
	IsEmailVerified bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// REQUEST

type UserSignUpRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}

// RESPONSE

type UserResponse struct {
	Id         int64  `json:"id"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	ProfilePic string `json:"profile_pic,omitempty"`
	Email      string `json:"email"`
}
