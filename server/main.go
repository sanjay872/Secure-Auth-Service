package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"

	"secure-auth-backend/database"
	"secure-auth-backend/internal/middleware"
	"secure-auth-backend/internal/user"
	"time"
)

func main() {

	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Read environment variables
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL not set")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET not set")
	}

	// Initialize Database
	db := database.NewPostgres(databaseURL)

	// Initialize Firebase
	// opt := option.WithCredentialsFile("firebase-service-account.json")
	// app, err := firebase.NewApp(context.Background(), nil, opt)
	// if err != nil {
	// 	log.Fatal("Failed to initialize Firebase:", err)
	// }

	// fbAuth, err := app.Auth(context.Background())
	// if err != nil {
	// 	log.Fatal("Failed to initialize Firebase Auth:", err)
	// }

	// // Initialize Auth Service
	// authService := auth.NewService(db, fbAuth, []byte(jwtSecret))
	// authHandler := auth.NewHandler(authService)

	// Setup Router
	r := chi.NewRouter()

	allowedOriginsEnv := os.Getenv("ALLOWED_ORIGINS")
	if allowedOriginsEnv == "" {
		log.Fatal("ALLOWED_ORIGINS not set")
	}

	allowedOrigins := strings.Split(allowedOriginsEnv, ",")

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
	}))

	rateLimiter := middleware.NewRateLimiter(5, time.Minute)

	// initialise all services and handlers

	// AUTH

	// USER
	userService := user.NewService(db)
	userHandler := user.NewHandler(userService)

	// Auth Routes

	// rate limiter - 5 request per min
	r.Group(func(r chi.Router) {
		r.Use(rateLimiter.Middleware)

		// r.Post("/auth/exchange", authHandler.Exchange)
		// r.Post("/auth/refresh", authHandler.Refresh)

		r.Post("/user", userHandler.CreateUser)
	})

	//r.Post("/auth/logout", authHandler.Logout)

	// Protected Routes
	r.Group(func(r chi.Router) {
		r.Use(rateLimiter.Middleware)
		//r.Use(authService.JWTMiddleware)

		//r.Get("/profile", authHandler.Profile)
	})

	log.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
