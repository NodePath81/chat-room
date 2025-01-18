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
	chimiddleware "github.com/go-chi/chi/v5/middleware"
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
		r.Get("/", sessionHandler.GetSessions)
		r.Post("/", sessionHandler.CreateSession)
		r.Get("/join", sessionHandler.JoinSession)
		r.Get("/share/info", sessionHandler.GetShareInfo)
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", sessionHandler.GetSession)
			r.Get("/role", sessionHandler.CheckRole)
			r.Get("/members", sessionHandler.ListMembers)
			r.Get("/kick", sessionHandler.KickMember)
			r.Get("/remove", sessionHandler.RemoveSession)
			r.Post("/share", sessionHandler.CreateShareLink)
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
