package server

import (
	"RoomBookingService/internal/auth"
	"RoomBookingService/internal/handler"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewServer(tokenService *auth.JWTService,
	userHandler *handler.UserHandler,
	roomHandler *handler.RoomHandler,
	scheduleHandler *handler.ScheduleHandler,
	slotHandler *handler.SlotHandler,
	bookingHandler *handler.BookingHandler,
) http.Handler {
	r := chi.NewRouter()

	// Глобальный middleware для всех маршрутов
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	authMiddleware := auth.NewAuthMiddleware(tokenService)

	// Публичные маршруты (без авторизации)
	r.Post("/dummyLogin", userHandler.DummyLogin)
	r.Get("/_info", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	r.Post("/register", userHandler.RegisterUserWithPassword)
	r.Post("/login", userHandler.Login)

	// Защищенные маршруты для admin
	r.Group(func(r chi.Router) {
		r.Use(authMiddleware.JWT)
		r.Use(auth.RequireRole("admin"))
		r.Get("/admin/ping", func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("admin ok"))
		})

		r.Post("/rooms/create", roomHandler.RegisterRoomRoutes)
		r.Post("/rooms/{roomId}/schedule/create", scheduleHandler.CreateSchedule)
		r.Get("/bookings/list", bookingHandler.GetAllBookings)
	})

	// Защищенные маршруты для user
	r.Group(func(r chi.Router) {
		r.Use(authMiddleware.JWT)
		r.Use(auth.RequireRole("user"))
		r.Get("/user/ping", func(w http.ResponseWriter, r *http.Request) {
			userID, _ := auth.UserIDFromContext(r.Context())
			role, _ := auth.RoleFromContext(r.Context())
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{"user_id": userID.String(), "role": role})
		})
		r.Post("/bookings/create", bookingHandler.CreateBooking)
		r.Get("/bookings/my", bookingHandler.GetMyBookings)
		r.Post("/bookings/{bookingId}/cancel", bookingHandler.CancelBooking)
	})

	// Защищенные маршруты для всех авторизованных пользователей
	r.Group(func(r chi.Router) {
		r.Use(authMiddleware.JWT)
		r.Get("/rooms/list", roomHandler.GetAllRooms)
		r.Get("/rooms/{roomId}/slots/list", slotHandler.GetAvailableSlots)
	})

	return r
}
