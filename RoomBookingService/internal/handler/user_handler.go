package handler

import (
	"RoomBookingService/internal/dto"
	"RoomBookingService/internal/httpresp"
	"RoomBookingService/internal/logger"
	"RoomBookingService/internal/usecase/service"
	myerrors "RoomBookingService/pkg/my_errors"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-playground/validator"
)

type UserHandler struct {
	service service.AuthService
}

func NewUserHandler(service service.AuthService) *UserHandler {
	return &UserHandler{service: service}
}

func (h *UserHandler) RegisterUserRoutes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)

	var req dto.CreateUserRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warn("register: invalid body", "error", err)
		httpresp.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid body")
		return
	}

	if role := req.Role; role != "user" && role != "admin" {
		log.Warn("register: invalid role", "role", role)
		httpresp.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid role")
		return
	}

	_, err := h.service.Register(ctx, req)
	if err != nil {
		if errors.Is(err, myerrors.ErrUserAlreadyExists) {
			log.Warn("register: user already exists", "username", req.Username)
			httpresp.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "user already exists")
			return
		}
		log.Error("register: service error", "error", err)
		httpresp.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}
	httpresp.WriteJSON(w, http.StatusCreated, map[string]string{"message": "user registered successfully"})

}

func (h *UserHandler) DummyLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)

	if r.Method != http.MethodPost {
		httpresp.WriteError(w, http.StatusMethodNotAllowed, "INVALID_REQUEST", "method not allowed")
		return
	}

	var req dto.DummyLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warn("dummy login: invalid body", "error", err)
		httpresp.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid body")
		return
	}

	req.Role = strings.ToLower(strings.TrimSpace(req.Role))
	if req.Role != "admin" && req.Role != "user" {
		log.Warn("dummy login: invalid role", "role", req.Role)
		httpresp.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid role")
		return
	}

	result, err := h.service.DummyLogin(ctx, req.Role)
	if err != nil {
		if errors.Is(err, myerrors.ErrInvalidRole) {
			log.Warn("dummy login: invalid role", "role", req.Role)
			httpresp.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid role")
			return
		}
		log.Error("dummy login: service error", "error", err)
		httpresp.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}
	httpresp.WriteJSON(w, http.StatusOK, result)
}

func (h *UserHandler) RegisterUserWithPassword(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	if r.Method != http.MethodPost {
		httpresp.WriteError(w, http.StatusMethodNotAllowed, "INVALID_REQUEST", "method not allowed")
		return
	}

	var req dto.CreateUserRequest
	var res *dto.AuthResult

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warn("register: invalid request", "error", err)
		httpresp.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request")
		return
	}

	res, err := h.service.RegisterWithPassword(r.Context(), req)
	if err != nil {
		var validationErr validator.ValidationErrors
		switch {
		case errors.Is(err, myerrors.ErrUserAlreadyExists):
			httpresp.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "user already exists")
		case errors.Is(err, myerrors.ErrInvalidRequest):
			httpresp.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request")
		case errors.As(err, &validationErr):
			httpresp.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request")
		default:
			log.Error("register: internal error", "error", err, "email", req.Email)
			httpresp.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		}
		return
	}

	httpresp.WriteJSON(w, http.StatusCreated, map[string]any{
		"user": map[string]any{
			"id":    res.UserID,
			"email": res.Email,
			"role":  res.Role,
		},
	})
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	if r.Method != http.MethodPost {
		httpresp.WriteError(w, http.StatusMethodNotAllowed, "INVALID_REQUEST", "method not allowed")
		return
	}

	var req dto.LoginRequest
	var res *dto.TokenResponse

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warn("login: invalid request", "error", err)
		httpresp.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request")
		return
	}

	res, err := h.service.Login(r.Context(), req)
	if err != nil {
		var validationErr validator.ValidationErrors
		switch {
		case errors.Is(err, myerrors.ErrInvalidCredentials):
			log.Warn("login: invalid credentials", "email", req.Email)
			httpresp.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid credentials")
		case errors.As(err, &validationErr):
			httpresp.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request")
		default:
			log.Error("login: internal error", "error", err, "email", req.Email)
			httpresp.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		}
		return
	}

	httpresp.WriteJSON(w, http.StatusOK, res)
}
