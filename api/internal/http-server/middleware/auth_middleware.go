package middleware

import (
	"badJokes/internal/config"
	"badJokes/internal/lib/sl"
	"context"
	"log/slog"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type key int

const (
	UserIDKey   key = iota
	UserAdminKey
)

type AuthMiddleware struct {
	jwtSecret []byte
	log       *slog.Logger
}

func NewAuthMiddleware(cfg *config.Config, log *slog.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		jwtSecret: []byte(cfg.JWTSecret),
		log:       log.With(slog.String("component", "auth_middleware")),
	}
}

func (a *AuthMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")

		if authHeader == "" {
			next.ServeHTTP(w, r)
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}

		tokenString := authHeader[len("Bearer "):]

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return a.jwtSecret, nil
		})

		if err != nil || !token.Valid {
			a.log.Debug("Invalid token", sl.Err(err))
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {
			a.log.Debug("Invalid token claims")
			http.Error(w, "Invalid token claims", http.StatusUnauthorized)
			return
		}

		userID, ok := claims["user_id"].(float64)
		if !ok {
			a.log.Debug("Invalid user ID in token")
			http.Error(w, "Invalid user ID in token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, int64(userID))
		
		if isAdmin, ok := claims["is_admin"].(bool); ok {
			ctx = context.WithValue(ctx, UserAdminKey, isAdmin)
			a.log.Debug("User authenticated", 
				slog.Int64("user_id", int64(userID)), 
				slog.Bool("is_admin", isAdmin))
		} else {
			ctx = context.WithValue(ctx, UserAdminKey, false)
			a.log.Debug("User authenticated (non-admin)", 
				slog.Int64("user_id", int64(userID)))
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (a *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(UserIDKey).(int64)
		if !ok {
			a.log.Info("Authentication required but not provided")
			http.Error(w, "Authentication required", http.StatusUnauthorized)
			return
		}

		a.log.Debug("Authenticated request", slog.Int64("user_id", userID))
		next.ServeHTTP(w, r)
	})
}

func (a *AuthMiddleware) RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(UserIDKey).(int64)
		if !ok {
			a.log.Info("Authentication required but not provided")
			http.Error(w, "Authentication required", http.StatusUnauthorized)
			return
		}

		isAdmin, ok := r.Context().Value(UserAdminKey).(bool)
		if !ok || !isAdmin {
			a.log.Info("Admin privileges required but not granted", 
				slog.Int64("user_id", userID))
			http.Error(w, "Forbidden: Admin privileges required", http.StatusForbidden)
			return
		}

		a.log.Debug("Admin request authorized", 
			slog.Int64("user_id", userID), 
			slog.Bool("is_admin", isAdmin))
		next.ServeHTTP(w, r)
	})
}