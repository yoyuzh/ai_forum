package user

import (
	"context"
	"errors"
	"testing"
)

func TestServiceRegisterHashesPasswordAndRejectsDuplicateUsername(t *testing.T) {
	repo := newMemoryRepository()
	svc := NewService(repo)
	ctx := context.Background()

	u, err := svc.Register(ctx, RegisterInput{Username: "alice", Password: "secret123", DisplayName: "Alice"})
	if err != nil {
		t.Fatal(err)
	}
	if u.ID == 0 {
		t.Fatal("expected user id")
	}
	if u.PasswordHash == "secret123" {
		t.Fatal("password stored in plaintext")
	}
	if !CheckPassword(u.PasswordHash, "secret123") {
		t.Fatal("bcrypt hash must verify original password")
	}

	_, err = svc.Register(ctx, RegisterInput{Username: "alice", Password: "another123"})
	if !errors.Is(err, ErrDuplicateUsername) {
		t.Fatalf("duplicate err = %v, want ErrDuplicateUsername", err)
	}
	if len(repo.users) != 1 {
		t.Fatalf("user count = %d, want 1", len(repo.users))
	}
}

func TestServiceProfileReturnsRegisteredUser(t *testing.T) {
	repo := newMemoryRepository()
	svc := NewService(repo)
	ctx := context.Background()
	u, err := svc.Register(ctx, RegisterInput{Username: "alice", Password: "secret123", DisplayName: "Alice"})
	if err != nil {
		t.Fatal(err)
	}

	profile, err := svc.Profile(ctx, u.ID)
	if err != nil {
		t.Fatal(err)
	}
	if profile.ID != u.ID || profile.Username != "alice" || profile.DisplayName != "Alice" {
		t.Fatalf("profile = %#v", profile)
	}
	if profile.PasswordHash != "" {
		t.Fatal("profile must not expose password hash")
	}
}
