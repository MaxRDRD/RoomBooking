package main

import (
	"RoomBookingService/cmd/server"
	"RoomBookingService/internal/auth"
	"RoomBookingService/internal/handler"
	"RoomBookingService/internal/logger"
	"RoomBookingService/internal/repository_impl/postgres"
	"RoomBookingService/internal/usecase/service"
	"context"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/go-playground/validator"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	log := logger.NewLogger()
	ctx := logger.WithContext(context.Background(), log)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// DATABASE_URL с дефолтом для локальной разработки
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://postgres:postgres@localhost:5432/room_booking"
	}

	// Подключение к БД
	pool, err := pgxpool.New(ctx, connStr)

	if err != nil {
		log.Error("failed to create database pool", "error", err,
			"stack", string(debug.Stack()))
		panic(err)
	}
	defer pool.Close()

	validator := validator.New()

	tokenService := auth.NewJWTService()

	userRepo := postgres.NewUserRepository(pool)
	roomRepo := postgres.NewRoomRepository(pool)
	scheduleRepo := postgres.NewScheduleRepository(pool)
	slotRepo := postgres.NewSlotRepository(pool)
	bookingRepo := postgres.NewBookingRepository(pool)

	userService := service.NewUserService(userRepo, validator, tokenService)
	roomService := service.NewRoomService(roomRepo)
	scheduleService := service.NewScheduleService(scheduleRepo)
	slotService := service.NewSlotService(slotRepo)
	bookingService := service.NewBookingService(bookingRepo)

	userHandler := handler.NewUserHandler(userService)
	roomHandler := handler.NewRoomHandler(roomService)
	scheduleHandler := handler.NewScheduleHandler(scheduleService, validator)
	slotHandler := handler.NewSlotHandler(*slotService)
	bookingHandler := handler.NewBookingHandler(bookingService, validator)
	// Создание роутера с зависимостями
	h := server.NewServer(tokenService, userHandler, roomHandler, scheduleHandler, slotHandler, bookingHandler)

	// Запуск сервера с graceful shutdown
	server := &http.Server{
		Addr:    ":" + port,
		Handler: h,
	}

	// Graceful shutdown в отдельной горутине
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		log.Info("shutting down server")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_ = server.Shutdown(ctx)
	}()

	log.Info("starting server", "port", port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Error("server error", "error", err)
		panic(err)
	}
}
