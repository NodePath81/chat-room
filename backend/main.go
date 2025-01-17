package main

import (
	"log"
	"net/http"

	"chat-room/config"
	"chat-room/database"
	"chat-room/handlers"
	custommw "chat-room/middleware"
	"chat-room/s3"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	// Initialize database
	db, err := database.InitDB(cfg)
	if err != nil {
		log.Fatal(err)
	}

	// Initialize MinIO
	err = s3.Initialize(cfg)
	if err != nil {
		log.Fatal("Failed to initialize MinIO:", err)
	}

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(db)
	sessionHandler := handlers.NewSessionHandler(db)
	wsHandler := handlers.NewWebSocketHandler(db)
	userHandler := handlers.NewUserHandler(db)
	avatarHandler := handlers.NewAvatarHandler(db)

	// Setup router
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
	}))

	// Routes
	r.Post("/api/auth/register", authHandler.Register)
	r.Post("/api/auth/login", authHandler.Login)
	r.Get("/api/auth/check-username", authHandler.CheckUsernameAvailability)
	r.Get("/api/auth/check-nickname", authHandler.CheckNicknameAvailability)

	// Session routes
	r.Group(func(r chi.Router) {
		r.Use(custommw.AuthMiddleware)
		r.Get("/api/sessions", sessionHandler.GetSessions)
		r.Post("/api/sessions", sessionHandler.CreateSession)
		r.Post("/api/sessions/{id}/join", sessionHandler.JoinSession)
		r.Get("/api/sessions/{id}/check", sessionHandler.CheckSessionMembership)
	})

	// User routes
	r.Group(func(r chi.Router) {
		r.Use(custommw.AuthMiddleware)
		r.Get("/api/users/{id}", userHandler.GetUser)
		r.Put("/api/users/{id}/nickname", userHandler.UpdateNickname)
		r.Put("/api/users/{id}/username", userHandler.UpdateUsername)
	})

	// Avatar routes
	r.Group(func(r chi.Router) {
		r.Use(custommw.AuthMiddleware)
		r.Post("/api/avatar", avatarHandler.UploadAvatar)
	})

	// WebSocket endpoint
	r.Get("/ws", wsHandler.HandleWebSocket)

	// Start server
	port := cfg.Port
	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
