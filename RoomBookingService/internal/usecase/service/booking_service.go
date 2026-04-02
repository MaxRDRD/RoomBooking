package service

import (
	"RoomBookingService/internal/dto"
	"RoomBookingService/internal/logger"
	"RoomBookingService/internal/model"
	"RoomBookingService/internal/repository"
	"context"

	"github.com/google/uuid"
)

type BookingService interface {
	CreateBooking(ctx context.Context, userID uuid.UUID, req dto.BookingRequest) (model.Booking, error)
	CancelBooking(ctx context.Context, userID uuid.UUID, req dto.CancelBookingRequest) (model.Booking, error)
	GetMyBookings(ctx context.Context, userID uuid.UUID, filter dto.BookingFilter) ([]model.Booking, error)
	GetAllBookings(ctx context.Context, filter dto.BookingFilter) ([]model.Booking, int, error)
}

type bookingService struct {
	bookingRepo repository.BookingRepository
}

func NewBookingService(bookingRepo repository.BookingRepository) BookingService {
	return &bookingService{bookingRepo: bookingRepo}
}

func (s *bookingService) CreateBooking(ctx context.Context, userID uuid.UUID, req dto.BookingRequest) (model.Booking, error) {
	log := logger.FromContext(ctx)
	booking, err := s.bookingRepo.CreateBooking(ctx, userID, req)
	if err != nil {
		log.Error("failed to create booking", "error", err)
		return model.Booking{}, err
	}
	return booking, nil
}

func (s *bookingService) CancelBooking(ctx context.Context, userID uuid.UUID, req dto.CancelBookingRequest) (model.Booking, error) {
	log := logger.FromContext(ctx)
	booking, err := s.bookingRepo.CancelBooking(ctx, userID, req)
	if err != nil {
		log.Error("failed to cancel booking", "error", err)
		return model.Booking{}, err
	}
	return booking, nil
}

func (s *bookingService) GetMyBookings(ctx context.Context, userID uuid.UUID, filter dto.BookingFilter) ([]model.Booking, error) {
	log := logger.FromContext(ctx)
	bookings, err := s.bookingRepo.GetMyBookings(ctx, userID, filter)
	if err != nil {
		log.Error("failed to get my bookings", "error", err)
		return nil, err
	}
	return bookings, nil
}

func (s *bookingService) GetAllBookings(ctx context.Context, filter dto.BookingFilter) ([]model.Booking, int, error) {
	log := logger.FromContext(ctx)
	bookings, total, err := s.bookingRepo.GetAllBookings(ctx, filter)
	if err != nil {
		log.Error("failed to get all bookings", "error", err)
		return nil, 0, err
	}
	return bookings, total, nil
}
