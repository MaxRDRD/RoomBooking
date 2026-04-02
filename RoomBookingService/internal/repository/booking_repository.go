package repository

import (
	"RoomBookingService/internal/dto"
	"RoomBookingService/internal/model"
	"context"

	"github.com/google/uuid"
)

type BookingRepository interface {
	CreateBooking(ctx context.Context, userID uuid.UUID, req dto.BookingRequest) (model.Booking, error)       // только user
	CancelBooking(ctx context.Context, userID uuid.UUID, req dto.CancelBookingRequest) (model.Booking, error) // только user
	GetMyBookings(ctx context.Context, userID uuid.UUID, filter dto.BookingFilter) ([]model.Booking, error)   // только user
	GetAllBookings(ctx context.Context, filter dto.BookingFilter) ([]model.Booking, int, error)               // только admin
}
