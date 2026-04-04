package auth

import (
	"RoomBookingService/internal/httpresp"
	myerrors "RoomBookingService/pkg/my_errors"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
)

func TestJWTService_GenerateAndParseDummyToken(t *testing.T) {
	service := NewJWTService()
	token, err := service.GenerateDummyToken("admin")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	userID, role, err := service.ParseToken(token)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if userID.String() != AdminDummyUserID {
		t.Fatalf("unexpected user id %s", userID)
	}
	if role != "admin" {
		t.Fatalf("unexpected role %s", role)
	}
}

func TestJWTService_InvalidRole(t *testing.T) {
	service := NewJWTService()
	if _, err := service.GenerateDummyToken("manager"); err != myerrors.ErrInvalidRole {
		t.Fatalf("expected invalid role error, got %v", err)
	}
}

func TestAuthMiddleware(t *testing.T) {
	service := NewJWTService()
	token, err := service.GenerateDummyToken("user")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, err := UserIDFromContext(r.Context())
		if err != nil {
			t.Fatalf("expected user id in context, got %v", err)
		}
		role, err := RoleFromContext(r.Context())
		if err != nil {
			t.Fatalf("expected role in context, got %v", err)
		}
		if userID == uuid.Nil || role != "user" {
			t.Fatalf("unexpected context payload: %s %s", userID, role)
		}
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp := httptest.NewRecorder()
	NewAuthMiddleware(service).JWT(handler).ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.Code)
	}
}

func TestAuthMiddleware_MissingHeader(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler should not run")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp := httptest.NewRecorder()
	NewAuthMiddleware(NewJWTService()).JWT(handler).ServeHTTP(resp, req)
	if resp.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.Code)
	}
}

func TestRequireRole(t *testing.T) {
	service := NewJWTService()
	token, err := service.GenerateDummyToken("admin")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	chain := NewAuthMiddleware(service).JWT(RequireRole("admin")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp := httptest.NewRecorder()
	chain.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.Code)
	}
}

func TestRoleFromContextMissing(t *testing.T) {
	_, err := RoleFromContext(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestUserIDFromContextMissing(t *testing.T) {
	_, err := UserIDFromContext(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestUnauthorizedErrorResponseShape(t *testing.T) {
	rec := httptest.NewRecorder()
	httpresp.WriteError(rec, http.StatusUnauthorized, "UNAUTHORIZED", "missing authorization header")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
	var body map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	errorPart, ok := body["error"].(map[string]any)
	if !ok || errorPart["code"] != "UNAUTHORIZED" {
		t.Fatalf("unexpected error body: %#v", body)
	}
}
