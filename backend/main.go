package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"chat-room/config"
	"chat-room/handlers"
	custommw "chat-room/middleware"
	"chat-room/s3"
	"chat-room/store/postgres"
	"chat-room/token"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	// Construct database URL
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
	)

	// Initialize store
	store, err := postgres.New(context.Background(), dbURL)
	if err != nil {
		log.Fatal("Failed to initialize store:", err)
	}
	defer store.Close()

	// Apply migrations
	if err := store.Migrate(context.Background()); err != nil {
		log.Fatal("Failed to apply migrations:", err)
	}

	// Initialize MinIO
	err = s3.Initialize(cfg)
	if err != nil {
		log.Fatal("Failed to initialize MinIO:", err)
	}

	// Initialize token manager
	tokenManager, err := token.NewManager(cfg.JWTSecret)
	if err != nil {
		log.Fatal("Failed to initialize token manager:", err)
	}

	// Initialize handlers
	wsHandler := handlers.NewWebSocketHandler(store, tokenManager)
	authHandler := handlers.NewAuthHandler(store)
	sessionHandler := handlers.NewSessionHandler(store, tokenManager)
	userHandler := handlers.NewUserHandler(store)
	avatarHandler := handlers.NewAvatarHandler(store)
	messageHandler := handlers.NewMessageHandler(store, wsHandler)

	// Setup router
	r := chi.NewRouter()

	// Middleware
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
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
	r.Route("/api/sessions", func(r chi.Router) {
		r.Use(custommw.AuthMiddleware)

		// Public session routes (require only auth)
		r.Get("/", sessionHandler.GetSessions)
		r.Post("/", sessionHandler.CreateSession)
		r.Get("/join", sessionHandler.JoinSession)
		r.Get("/share/info", sessionHandler.GetShareInfo)
		r.Get("/token", sessionHandler.GetSessionToken)

		// Protected session routes requiring session token
		r.Group(func(r chi.Router) {
			r.Use(custommw.NewSessionAuth(tokenManager))

			// Basic session member routes
			r.Get("/session", sessionHandler.GetSession)
			r.Get("/role", sessionHandler.CheckRole)
			r.Get("/members", sessionHandler.ListMembers)
			r.Get("/messages", sessionHandler.GetMessages)
			r.Post("/messages/upload", messageHandler.UploadMessageImage)
			r.Get("/token/refresh", sessionHandler.RefreshSessionToken)
			r.Delete("/token", sessionHandler.RevokeSessionToken)
			r.Get("/wstoken", sessionHandler.GetWebSocketToken)
			// Creator-only routes
			r.Group(func(r chi.Router) {
				r.Use(custommw.RequireRole("creator"))
				r.Get("/kick", sessionHandler.KickMember)
				r.Get("/remove", sessionHandler.RemoveSession)
				r.Post("/share", sessionHandler.CreateShareLink)
			})
		})
	})

	// User routes
	r.Group(func(r chi.Router) {
		r.Use(custommw.AuthMiddleware)
		r.Get("/api/users/{id}", userHandler.GetUser)
		r.Get("/api/users/sessions", userHandler.GetUserSessions)
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
