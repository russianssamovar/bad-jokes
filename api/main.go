package main

import (
	"badJokes/internal/config"
	"badJokes/internal/http-server/handlers"
	"badJokes/internal/http-server/middleware"
	"badJokes/internal/lib/sl"
	"badJokes/internal/storage"
	"database/sql"
	"flag"
	"log/slog"
	"net/http"
	"os"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	cfg := config.MustLoad()
	log := setupLogger(cfg.Env)

	db, err := sql.Open(cfg.Db.Driver, cfg.Db.ConnectionString)
	if err != nil {
		log.Error("Failed to connect to database", sl.Err(err))
		os.Exit(1)
	}
	defer db.Close()

	if err := storage.Migrate(db, "storage/migrations"); err != nil {
		log.Error("Failed to run migrations", sl.Err(err))
		os.Exit(1)
	}

	userRepo := storage.NewUserRepository(cfg.Db.Driver, db)
	jokesRepo := storage.NewJokesRepository(cfg.Db.Driver, db)

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
	switch env {
	case "local":
		return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case "prod":
		return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	default:
		return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	}
}
