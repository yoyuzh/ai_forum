// Package auth owns user JWT handling and authenticated subjects.
package auth

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

type Subject struct {
	UserID   int64
	Username string
	Role     string
}

type subjectKey struct{}

func SubjectFromContext(ctx context.Context) (Subject, bool) {
	sub, ok := ctx.Value(subjectKey{}).(Subject)
	return sub, ok
}

func ContextWithSubject(ctx context.Context, sub Subject) context.Context {
	return context.WithValue(ctx, subjectKey{}, sub)
}

type TokenManager struct {
	secret []byte
	ttl    time.Duration
	now    func() time.Time
}

type Authenticator interface {
	Authenticate(context.Context, string, string) (Subject, error)
}

type Handler struct {
	authenticator Authenticator
	tokens        *TokenManager
}

func NewHandler(authenticator Authenticator, tokens *TokenManager) *Handler {
	return &Handler{authenticator: authenticator, tokens: tokens}
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	sub, err := h.authenticator.Authenticate(r.Context(), req.Username, req.Password)
	if err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}
	token, err := h.tokens.Issue(sub)
	if err != nil {
		http.Error(w, "issue token", http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]string{"token": token})
}

func NewTokenManager(secret string, ttl time.Duration) *TokenManager {
	return &TokenManager{secret: []byte(secret), ttl: ttl, now: time.Now}
}

func (m *TokenManager) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		token, ok := strings.CutPrefix(header, "Bearer ")
		if !ok || token == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		sub, err := m.Validate(token)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r.WithContext(ContextWithSubject(r.Context(), sub)))
	})
}

func (m *TokenManager) Issue(sub Subject) (string, error) {
	header := map[string]string{"alg": "HS256", "typ": "JWT"}
	claims := map[string]any{
		"sub":      strconv.FormatInt(sub.UserID, 10),
		"username": sub.Username,
		"role":     sub.Role,
		"exp":      m.now().Add(m.ttl).Unix(),
	}
	head, err := encodeJSON(header)
	if err != nil {
		return "", err
	}
	body, err := encodeJSON(claims)
	if err != nil {
		return "", err
	}
	unsigned := head + "." + body
	return unsigned + "." + m.sign(unsigned), nil
}

func (m *TokenManager) Validate(token string) (Subject, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return Subject{}, errors.New("malformed token")
	}
	unsigned := parts[0] + "." + parts[1]
	if !hmac.Equal([]byte(parts[2]), []byte(m.sign(unsigned))) {
		return Subject{}, errors.New("invalid token signature")
	}
	var claims struct {
		Sub      string `json:"sub"`
		Username string `json:"username"`
		Role     string `json:"role"`
		Exp      int64  `json:"exp"`
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return Subject{}, err
	}
	if err := json.Unmarshal(raw, &claims); err != nil {
		return Subject{}, err
	}
	if m.now().Unix() >= claims.Exp {
		return Subject{}, errors.New("token expired")
	}
	userID, err := strconv.ParseInt(claims.Sub, 10, 64)
	if err != nil {
		return Subject{}, fmt.Errorf("invalid subject: %w", err)
	}
	return Subject{UserID: userID, Username: claims.Username, Role: claims.Role}, nil
}

func (m *TokenManager) sign(unsigned string) string {
	mac := hmac.New(sha256.New, m.secret)
	_, _ = mac.Write([]byte(unsigned))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func encodeJSON(v any) (string, error) {
	raw, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}
