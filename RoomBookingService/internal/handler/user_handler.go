package handler

import (
	"RoomBookingService/internal/dto"
	"RoomBookingService/internal/logger"
	"RoomBookingService/internal/usecase"
	"encoding/json"
	"net/http"
	"strings"
)

type UserHandler struct {
	service usecase.AuthService
}

func NewUserHandler(service usecase.AuthService) *UserHandler {
	return &UserHandler{service: service}
}

func (h *UserHandler) RegisterUserRoutes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)

	var req dto.CreateUserRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warn("register: invalid body", "error", err)
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	if role := req.Role; role != "user" && role != "admin" {
		log.Warn("register: invalid role", "role", role)
		http.Error(w, "invalid role", http.StatusBadRequest)
		return
	}

	result, err := h.service.Register(ctx, req)
	if err != nil {
		log.Error("register: service error", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)

}

func (h *UserHandler) DummyLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req dto.DummyLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warn("dummy login: invalid body", "error", err)
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	req.Role = strings.ToLower(strings.TrimSpace(req.Role))
	if req.Role != "admin" && req.Role != "user" {
		log.Warn("dummy login: invalid role", "role", req.Role)
		http.Error(w, "invalid role", http.StatusBadRequest)
		return
	}

	resp, err := h.service.DummyLogin(ctx, req.Role)
	if err != nil {
		log.Error("dummy login: service error", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}
