package main

import (
	"badJokes/api/internal/config"
	"badJokes/api/internal/http-server/handlers"
	"badJokes/api/internal/http-server/middleware"
	"badJokes/api/internal/lib/sl"
	"badJokes/api/internal/storage/sqlite/jokes"
	"badJokes/api/internal/storage/sqlite/user"
	"database/sql"
	"flag"
	"log/slog"
	"net/http"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	cfg := config.MustLoad()
	log := setupLogger(cfg.Env)

	log.Info("Starting bad-jokes app", slog.String("env", cfg.Env))

	db, err := sql.Open(cfg.Db.Driver, cfg.Db.ConnectionString)
	if err != nil {
		log.Error("Failed to connect to database", sl.Err(err))
		os.Exit(1)
	}
	defer db.Close()

	jokesRepo := jokes.NewJokesRepository(db)
	userRepo := user.NewUserRepository(db)

	jokesHandler := handlers.NewJokesHandler(jokesRepo)
	authHandler := handlers.NewAuthHandler(userRepo, cfg)

	authMiddleware := middleware.NewAuthMiddleware(cfg)

	mux := http.NewServeMux()
	setupRoutes(mux, jokesHandler, authHandler, authMiddleware)
	handler := corsMiddleware(mux)

	listenAddr := flag.String("listenaddr", cfg.HTTPServer.Address, "HTTP server listen address")
	flag.Parse()

	log.Info("Server started at", slog.String("address", cfg.HTTPServer.Address))
	if err := http.ListenAndServe(*listenAddr, handler); err != nil {
		log.Error("Failed to start server", sl.Err(err))
		os.Exit(1)
	}
}

func setupRoutes(mux *http.ServeMux, jokesHandler *handlers.JokesHandler, authHandler *handlers.AuthHandler, authMiddleware *middleware.AuthMiddleware) {
	mux.HandleFunc("/api/auth/register", authHandler.Register)
	mux.HandleFunc("/api/auth/login", authHandler.Login)

	mux.Handle("/api/jokes", authMiddleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			jokesHandler.ListJokes(w, r)
		case http.MethodPost:
			jokesHandler.CreateJoke(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	mux.Handle("/api/jokes/vote", authMiddleware.Middleware(http.HandlerFunc(jokesHandler.VoteEntity)))
	mux.Handle("/api/jokes/react", authMiddleware.Middleware(http.HandlerFunc(jokesHandler.ReactToEntity)))
	mux.Handle("/api/jokes/delete", authMiddleware.Middleware(http.HandlerFunc(jokesHandler.DeleteJoke)))

	mux.Handle("/api/comments", authMiddleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			jokesHandler.ListComments(w, r)
		case http.MethodPost:
			jokesHandler.AddComment(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	mux.Handle("/api/comments/vote", authMiddleware.Middleware(http.HandlerFunc(jokesHandler.VoteEntity)))
	mux.Handle("/api/comments/react", authMiddleware.Middleware(http.HandlerFunc(jokesHandler.ReactToEntity)))
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case "local":
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case "prod":
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	default:
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	}

	return log
}
