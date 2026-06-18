package tests

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/arilbois/contentbank-v2/internal/models"
	"github.com/arilbois/contentbank-v2/internal/services/auth"
)

func TestAuth_HashAndCheckPassword(t *testing.T) {
	svc := auth.NewService(nil, "test-secret", time.Hour)
	hash, err := svc.HashPassword("super-secret-password")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	if hash == "" || hash == "super-secret-password" {
		t.Fatalf("hash should not be empty or equal plaintext")
	}
	if err := svc.CheckPassword(hash, "super-secret-password"); err != nil {
		t.Fatalf("check correct password: %v", err)
	}
	if err := svc.CheckPassword(hash, "wrong"); err == nil {
		t.Fatalf("expected error for wrong password")
	}
}

func TestAuth_RejectShortPassword(t *testing.T) {
	svc := auth.NewService(nil, "test-secret", time.Hour)
	if _, err := svc.HashPassword("short"); err == nil {
		t.Fatalf("expected error for short password")
	}
}

func TestAuth_TokenRoundTrip(t *testing.T) {
	svc := auth.NewService(nil, "test-secret", time.Hour)
	u := &models.User{
		ID:       uuid.New(),
		Username: "admin",
		Role:     "admin",
	}
	tok, err := svc.GenerateToken(u)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	claims, err := svc.ValidateToken(tok)
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	if claims.Username != "admin" || claims.Role != "admin" {
		t.Fatalf("claims mismatch: %+v", claims)
	}
	if claims.UserID != u.ID.String() {
		t.Fatalf("user id mismatch: got %s want %s", claims.UserID, u.ID.String())
	}
}

func TestAuth_TokenRejectsTampering(t *testing.T) {
	svc := auth.NewService(nil, "test-secret", time.Hour)
	u := &models.User{ID: uuid.New(), Username: "viewer", Role: "viewer"}
	tok, _ := svc.GenerateToken(u)
	bad := tok + "tamper"
	if _, err := svc.ValidateToken(bad); err == nil {
		t.Fatalf("expected tampered token to fail validation")
	}
}

func TestAuth_TokenRejectsWrongSecret(t *testing.T) {
	signer := auth.NewService(nil, "secret-A", time.Hour)
	verifier := auth.NewService(nil, "secret-B", time.Hour)
	u := &models.User{ID: uuid.New(), Username: "u", Role: "viewer"}
	tok, _ := signer.GenerateToken(u)
	if _, err := verifier.ValidateToken(tok); err == nil {
		t.Fatalf("expected token signed with secret-A to fail verification with secret-B")
	}
}
