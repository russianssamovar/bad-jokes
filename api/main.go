package main

import (
	"badJokes/internal/config"
	"badJokes/internal/http-server/handlers"
	"badJokes/internal/http-server/middleware"
	"badJokes/internal/lib/sl"
	"badJokes/internal/storage"
	"context"
	"database/sql"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"strings"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	cfg := config.MustLoad()
	log := setupLogger(cfg.Env)

	log.Info("Starting application", slog.String("env", cfg.Env))

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
	log.Info("Database migrations completed successfully")

	userRepo := storage.NewUserRepository(cfg.Db.Driver, db, log)
	jokesRepo := storage.NewJokesRepository(cfg.Db.Driver, db, log)
	commentRepo := storage.NewCommentsRepository(cfg.Db.Driver, db, log)
	entityRepo := storage.NewEntityRepository(cfg.Db.Driver, db, log)

	jokesHandler := handlers.NewJokesHandler(jokesRepo, commentRepo, log)
	commentHandler := handlers.NewCommentHandler(commentRepo, log)
	entityHandler := handlers.NewEntityHandler(entityRepo, log)
	authHandler := handlers.NewAuthHandler(userRepo, cfg, log)
	oauthHandler := handlers.NewOAuthHandler(userRepo, cfg, log)
	adminHandler := handlers.NewAdminHandler(userRepo, jokesRepo, commentRepo, log)

	authMiddleware := middleware.NewAuthMiddleware(cfg, log)

	mux := http.NewServeMux()
	setupRoutes(mux, jokesHandler, commentHandler, entityHandler, authHandler, adminHandler, oauthHandler, authMiddleware)
	handler := corsMiddleware(mux)

	listenAddr := flag.String("listenaddr", cfg.HTTPServer.Address, "HTTP server listen address")
	flag.Parse()

	log.Info("Server started", slog.String("address", cfg.HTTPServer.Address))
	if err := http.ListenAndServe(*listenAddr, handler); err != nil {
		log.Error("Failed to start server", sl.Err(err))
		os.Exit(1)
	}
}

func setupRoutes(
	mux *http.ServeMux,
	jokesHandler *handlers.JokesHandler,
	commentHandler *handlers.CommentHandler,
	entityHandler *handlers.EntityHandler,
	authHandler *handlers.AuthHandler,
	adminHandler *handlers.AdminHandler,
	oauthHandler *handlers.OAuthHandler,
	authMiddleware *middleware.AuthMiddleware,
) {
	mux.HandleFunc("/api/auth/register", authHandler.Register)
	mux.HandleFunc("/api/auth/login", authHandler.Login)

	// OAuth routes
	mux.HandleFunc("/api/auth/google/login", func(w http.ResponseWriter, r *http.Request) {
		oauthHandler.InitiateOAuth(w, r, "google")
	})
	mux.HandleFunc("/api/auth/google/callback", func(w http.ResponseWriter, r *http.Request) {
		oauthHandler.OAuthCallback(w, r, "google")
	})
	mux.HandleFunc("/api/auth/github/login", func(w http.ResponseWriter, r *http.Request) {
		oauthHandler.InitiateOAuth(w, r, "github")
	})
	mux.HandleFunc("/api/auth/github/callback", func(w http.ResponseWriter, r *http.Request) {
		oauthHandler.OAuthCallback(w, r, "github")
	})

	mux.Handle("/api/jokes", authMiddleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			jokesHandler.List(w, r)
		case http.MethodPost:
			jokesHandler.Create(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	mux.Handle("/api/votes", authMiddleware.Middleware(http.HandlerFunc(entityHandler.Vote)))
	mux.Handle("/api/reactions", authMiddleware.Middleware(http.HandlerFunc(entityHandler.HandleReaction)))

	mux.Handle("/api/jokes/react", authMiddleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			entityHandler.HandleReaction(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	mux.Handle("/api/jokes/vote", authMiddleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			entityHandler.Vote(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	mux.Handle("/api/jokes/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		pathSegments := strings.Split(strings.TrimPrefix(path, "/api/jokes/"), "/")

		if len(pathSegments) == 0 || pathSegments[0] == "" {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}

		jokeID := pathSegments[0]

		if len(pathSegments) == 1 {
			r = r.WithContext(context.WithValue(r.Context(), "jokeId", jokeID))

			switch r.Method {
			case http.MethodGet:
				authMiddleware.Middleware(http.HandlerFunc(jokesHandler.GetJokeWithComments)).ServeHTTP(w, r)
			case http.MethodDelete:
				authMiddleware.Middleware(http.HandlerFunc(jokesHandler.DeleteJoke)).ServeHTTP(w, r)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
			return
		}

		if len(pathSegments) == 2 && pathSegments[1] == "comments" {
			r = r.WithContext(context.WithValue(r.Context(), "jokeId", jokeID))

			switch r.Method {
			case http.MethodPost:
				authMiddleware.Middleware(http.HandlerFunc(commentHandler.AddComment)).ServeHTTP(w, r)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
			return
		}

		http.Error(w, "Not found", http.StatusNotFound)
	}))

	mux.Handle("/api/comments", authMiddleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			commentHandler.ListComments(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	mux.Handle("/api/comments/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		pathSegments := strings.Split(strings.TrimPrefix(path, "/api/comments/"), "/")

		if len(pathSegments) != 1 || pathSegments[0] == "" {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}

		commentID := pathSegments[0]
		r = r.WithContext(context.WithValue(r.Context(), "commentId", commentID))

		switch r.Method {
		case http.MethodDelete:
			authMiddleware.Middleware(http.HandlerFunc(commentHandler.DeleteComment)).ServeHTTP(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	mux.Handle("/api/admin/users", authMiddleware.Middleware(
		authMiddleware.RequireAdmin(http.HandlerFunc(adminHandler.GetUsers)),
	))

	mux.Handle("/api/admin/jokes/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		pathSegments := strings.Split(strings.TrimPrefix(path, "/api/admin/jokes/"), "/")

		if len(pathSegments) != 1 || pathSegments[0] == "" {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}

		jokeID := pathSegments[0]
		r = r.WithContext(context.WithValue(r.Context(), "jokeId", jokeID))

		switch r.Method {
		case http.MethodDelete:
			authMiddleware.Middleware(
				authMiddleware.RequireAdmin(http.HandlerFunc(adminHandler.DeleteJoke)),
			).ServeHTTP(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	mux.Handle("/api/admin/comments/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		pathSegments := strings.Split(strings.TrimPrefix(path, "/api/admin/comments/"), "/")

		if len(pathSegments) != 1 || pathSegments[0] == "" {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}

		commentID := pathSegments[0]
		r = r.WithContext(context.WithValue(r.Context(), "commentId", commentID))

		switch r.Method {
		case http.MethodDelete:
			authMiddleware.Middleware(
				authMiddleware.RequireAdmin(http.HandlerFunc(adminHandler.DeleteComment)),
			).ServeHTTP(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	mux.Handle("/api/admin/users/set-status", authMiddleware.RequireAdmin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			adminHandler.SetUserAdminStatus(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	mux.Handle("/api/admin/logs", authMiddleware.Middleware(
		authMiddleware.RequireAdmin(http.HandlerFunc(adminHandler.GetModLogs)),
	))
	mux.Handle("/api/admin/stats", authMiddleware.Middleware(
		authMiddleware.RequireAdmin(http.HandlerFunc(adminHandler.GetUserStats)),
	))
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
