package handlers

import (
	"badJokes/internal/config"
	"badJokes/internal/lib/sl"
	"badJokes/internal/storage"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
)

type OAuthHandler struct {
	userRepo   storage.UserRepository
	log        *slog.Logger
	jwtSecret  []byte
	oauthConfs map[string]*oauth2.Config
	config     *config.Config
}

func NewOAuthHandler(repo storage.UserRepository, cfg *config.Config, log *slog.Logger) *OAuthHandler {
	googleConf := &oauth2.Config{
		ClientID:     cfg.OAuth.GoogleClientID,
		ClientSecret: cfg.OAuth.GoogleClientSecret,
		RedirectURL:  cfg.OAuth.BaseURL + "/api/auth/google/callback",
		Scopes:       []string{"profile", "email"},
		Endpoint:     google.Endpoint,
	}

	githubConf := &oauth2.Config{
		ClientID:     cfg.OAuth.GithubClientID,
		ClientSecret: cfg.OAuth.GithubClientSecret,
		RedirectURL:  cfg.OAuth.BaseURL + "/api/auth/github/callback",
		Scopes:       []string{"user:email"},
		Endpoint:     github.Endpoint,
	}

	return &OAuthHandler{
		userRepo:  repo,
		log:       log.With(slog.String("component", "oauth_handler")),
		jwtSecret: []byte(cfg.JWTSecret),
		oauthConfs: map[string]*oauth2.Config{
			"google": googleConf,
			"github": githubConf,
		},
		config: cfg,
	}
}

func (h *OAuthHandler) InitiateOAuth(w http.ResponseWriter, r *http.Request, provider string) {
	h.log.Debug("OAuth login initiated", slog.String("provider", provider))

	conf, ok := h.oauthConfs[provider]
	if !ok {
		http.Error(w, "Unsupported OAuth provider", http.StatusBadRequest)
		return
	}

	state := generateRandomState()

	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		MaxAge:   int(time.Hour.Seconds()),
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})

	url := conf.AuthCodeURL(state)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (h *OAuthHandler) OAuthCallback(w http.ResponseWriter, r *http.Request, provider string) {
	h.log.Debug("OAuth callback received", slog.String("provider", provider))

	stateCookie, err := r.Cookie("oauth_state")
	if err != nil || stateCookie.Value != r.FormValue("state") {
		h.log.Error("Invalid OAuth state", sl.Err(err))
		http.Error(w, "Invalid OAuth state", http.StatusBadRequest)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:   "oauth_state",
		Value:  "",
		MaxAge: -1,
		Path:   "/",
	})

	conf, ok := h.oauthConfs[provider]
	if !ok {
		http.Error(w, "Unsupported OAuth provider", http.StatusBadRequest)
		return
	}

	code := r.FormValue("code")
	token, err := conf.Exchange(context.Background(), code)
	if err != nil {
		h.log.Error("Failed to exchange OAuth code for token", sl.Err(err))
		http.Error(w, "Failed to complete OAuth flow", http.StatusInternalServerError)
		return
	}

	userInfo, err := h.getUserInfoFromProvider(provider, token)
	if err != nil {
		h.log.Error("Failed to get user info from provider",
			sl.Err(err),
			slog.String("provider", provider))
		http.Error(w, "Failed to get user information", http.StatusInternalServerError)
		return
	}

	user, err := h.userRepo.FindOrCreateOAuthUser(
		userInfo.Email,
		userInfo.Name,
		provider,
		userInfo.ProviderID,
	)
	if err != nil {
		h.log.Error("Failed to create or find user", sl.Err(err))
		http.Error(w, "Failed to process user data", http.StatusInternalServerError)
		return
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"is_admin": user.IsAdmin,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := jwtToken.SignedString(h.jwtSecret)
	if err != nil {
		h.log.Error("Failed to generate token", sl.Err(err))
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	callbackURL := h.config.OAuth.CallbackURL + "/auth/callback?token=" + tokenString
	http.Redirect(w, r, callbackURL, http.StatusTemporaryRedirect)
}

type OAuthUserInfo struct {
	Email      string
	Name       string
	ProviderID string
}

func (h *OAuthHandler) getUserInfoFromProvider(provider string, token *oauth2.Token) (*OAuthUserInfo, error) {
	client := h.oauthConfs[provider].Client(context.Background(), token)

	switch provider {
	case "google":
		resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		var userInfo struct {
			ID    string `json:"id"`
			Email string `json:"email"`
			Name  string `json:"name"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
			return nil, err
		}
        
        username := extractUsernameFromEmail(userInfo.Email)
		
		return &OAuthUserInfo{
			Email:      userInfo.Email,
			Name:       username,
			ProviderID: userInfo.ID,
		}, nil

	case "github":
		resp, err := client.Get("https://api.github.com/user")
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		var userInfo struct {
			ID    int    `json:"id"`
			Login string `json:"login"`
			Name  string `json:"name"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
			return nil, err
		}

		emailResp, err := client.Get("https://api.github.com/user/emails")
		if err != nil {
			return nil, err
		}
		defer emailResp.Body.Close()

		var emails []struct {
			Email    string `json:"email"`
			Primary  bool   `json:"primary"`
			Verified bool   `json:"verified"`
		}

		if err := json.NewDecoder(emailResp.Body).Decode(&emails); err != nil {
			return nil, err
		}

		var email string
		for _, e := range emails {
			if e.Primary && e.Verified {
				email = e.Email
				break
			}
		}

		username := extractUsernameFromEmail(email)

		return &OAuthUserInfo{
			Email:      email,
			Name:       username,
			ProviderID: string(userInfo.ID),
		}, nil
	}

	return nil, nil
}

func generateRandomState() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func extractUsernameFromEmail(email string) string {
	username := ""
	for i := 0; i < len(email); i++ {
		if email[i] == '@' {
			username = email[:i]
			break
		}
	}
	return username
}