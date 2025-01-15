package main

import (
	"log"
	"net/http"

	"chat-room/config"
	"chat-room/database"
	"chat-room/handlers"
	custommw "chat-room/middleware"

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

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(db)
	sessionHandler := handlers.NewSessionHandler(db)
	wsHandler := handlers.NewWebSocketHandler(db)

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
	r.HandleFunc("/ws", wsHandler.HandleWebSocket)

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(custommw.AuthMiddleware)

		r.Route("/api/sessions", func(r chi.Router) {
			r.Get("/", sessionHandler.GetSessions)
			r.Post("/", sessionHandler.CreateSession)
			r.Get("/{id}", sessionHandler.GetSession)
		})

	})

	// Start server
	port := cfg.Port
	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
