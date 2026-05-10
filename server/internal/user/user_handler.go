package user

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"secure-auth-backend/model"
)

type UserHandler struct {
	service *UserService
}

func NewHandler(_service *UserService) *UserHandler {
	return &UserHandler{service: _service}
}

func (uh *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {

	ctx := context.Background()

	var UserSignUpRequest model.UserSignUpRequest

	err := json.NewDecoder(r.Body).Decode(&UserSignUpRequest)

	// fmt.Print(UserSignUpRequest)

	if err != nil {
		http.Error(w, "invalid body request", http.StatusBadRequest)
		return
	}

	createdUser, err := uh.service.createUser(ctx, UserSignUpRequest)

	if err != nil {
		http.Error(w, "Failed to insert the data", http.StatusInternalServerError)
		fmt.Print(err)
		return
	}

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]any{
		"message": "User Created Successfully",
		"user":    createdUser,
	})
}
