package server

import (
	"RoomBookingService/internal/auth"
	"RoomBookingService/internal/handler"
	"RoomBookingService/internal/usecase"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/validator"
)

func NewServer(tokenService *auth.JWTService,
	userService usecase.AuthService,
	userHandler *handler.UserHandler,
) http.Handler {
	if tokenService == nil {
		tokenService = auth.NewJWTService()
	}
	if userService == nil {
		userService = usecase.NewUserService(nil, validator.New(), tokenService)
	}
	if userHandler == nil {
		userHandler = handler.NewUserHandler(userService)
	}

	r := chi.NewRouter()

	// Глобальный middleware для всех маршрутов
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// Публичные маршруты (без авторизации)
	r.Post("/dummyLogin", userHandler.DummyLogin)
	r.Get("/_info", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Защищенные маршруты для admin
	r.Group(func(r chi.Router) {
		r.Use(auth.AuthMiddleware(tokenService))
		r.Use(auth.RequireRole("admin"))
		r.Get("/admin/ping", func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("admin ok"))
		})
	})

	// Защищенные маршруты для user
	r.Group(func(r chi.Router) {
		r.Use(auth.AuthMiddleware(tokenService))
		r.Use(auth.RequireRole("user"))
		r.Get("/user/ping", func(w http.ResponseWriter, r *http.Request) {
			userID, _ := auth.UserIDFromContext(r.Context())
			role, _ := auth.RoleFromContext(r.Context())
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{"user_id": userID.String(), "role": role})
		})
	})

	return r
}
