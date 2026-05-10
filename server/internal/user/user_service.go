package user

import (
	"context"
	"fmt"
	"log"
	"secure-auth-backend/model"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	DB *pgxpool.Pool
}

func NewService(db *pgxpool.Pool) *UserService {
	return &UserService{DB: db}
}

func (s *UserService) createUser(ctx context.Context, user model.UserSignUpRequest) (*model.User, error) {

	fmt.Print("Got in!")

	query := `
		INSERT INTO users(first_name, last_name, email, password_hash)
		VALUES($1, $2, $3, $4)
		RETURNING id, first_name, last_name, email, profile_pic, created_at, updated_at
	`

	hashedPassword, err := hashPassword(user.Password)

	if err != nil {
		return nil, err
	}

	var createdUser model.User

	errFromDB := s.DB.QueryRow(ctx, query, user.FirstName, user.LastName, user.Email, hashedPassword).Scan(
		&createdUser.Id,
		&createdUser.FirstName,
		&createdUser.LastName,
		&createdUser.Email,
		&createdUser.ProfilePic,
		&createdUser.CreatedAt,
		&createdUser.UpdatedAt,
	)

	if errFromDB != nil {
		log.Print(errFromDB)
		return &model.User{}, errFromDB
	}

	fmt.Print(createdUser)

	return &createdUser, nil
}

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		return "", err
	}

	return string(hash), nil
}

func comparePassword(password string, existingHashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(existingHashedPassword), []byte(password))
	return err != nil
}
