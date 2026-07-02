// Package user owns user account domain behavior.
package user

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"ai-forum/backend/internal/auth"
	"golang.org/x/crypto/bcrypt"
)

var ErrDuplicateUsername = errors.New("duplicate username")

type User struct {
	ID           int64  `db:"id"`
	Username     string `db:"username"`
	PasswordHash string `db:"password_hash"`
	Role         string `db:"role"`
	DisplayName  string `db:"display_name"`
	Status       string `db:"status"`
}

type RegisterInput struct {
	Username    string
	Password    string
	DisplayName string
}

type Repository interface {
	Create(ctx context.Context, u User) (User, error)
	FindByID(ctx context.Context, id int64) (User, error)
	FindByUsername(ctx context.Context, username string) (User, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Register(ctx context.Context, in RegisterInput) (User, error) {
	username := strings.TrimSpace(in.Username)
	if username == "" || len(in.Password) < 8 {
		return User{}, fmt.Errorf("invalid registration")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, err
	}
	return s.repo.Create(ctx, User{
		Username:     username,
		PasswordHash: string(hash),
		Role:         "USER",
		DisplayName:  strings.TrimSpace(in.DisplayName),
		Status:       "ACTIVE",
	})
}

func (s *Service) Profile(ctx context.Context, id int64) (User, error) {
	u, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return User{}, err
	}
	u.PasswordHash = ""
	return u, nil
}

func (s *Service) Authenticate(ctx context.Context, username, password string) (auth.Subject, error) {
	u, err := s.repo.FindByUsername(ctx, strings.TrimSpace(username))
	if err != nil || !CheckPassword(u.PasswordHash, password) {
		return auth.Subject{}, auth.ErrInvalidCredentials
	}
	return auth.Subject{UserID: u.ID, Username: u.Username, Role: u.Role}, nil
}

func CheckPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

type memoryRepository struct {
	mu     sync.Mutex
	nextID int64
	users  map[string]User
}

func newMemoryRepository() *memoryRepository {
	return &memoryRepository{nextID: 1, users: map[string]User{}}
}

func (r *memoryRepository) Create(_ context.Context, u User) (User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.users[u.Username]; exists {
		return User{}, ErrDuplicateUsername
	}
	u.ID = r.nextID
	r.nextID++
	r.users[u.Username] = u
	return u, nil
}

func (r *memoryRepository) FindByID(_ context.Context, id int64) (User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, u := range r.users {
		if u.ID == id {
			return u, nil
		}
	}
	return User{}, fmt.Errorf("user not found")
}

func (r *memoryRepository) FindByUsername(_ context.Context, username string) (User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	u, ok := r.users[username]
	if !ok {
		return User{}, fmt.Errorf("user not found")
	}
	return u, nil
}
