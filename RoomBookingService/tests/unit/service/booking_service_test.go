package unit_test

import (
	"RoomBookingService/internal/dto"
	"RoomBookingService/internal/model"
	"RoomBookingService/internal/usecase/service"
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

type bookingRepoStub struct {
	createResult model.Booking
	createErr    error

	allResult []model.Booking
	allTotal  int
	allErr    error
}

func (b *bookingRepoStub) CreateBooking(_ context.Context, _ uuid.UUID, _ dto.BookingRequest) (model.Booking, error) {
	return b.createResult, b.createErr
}

func (b *bookingRepoStub) CancelBooking(_ context.Context, _ uuid.UUID, _ dto.CancelBookingRequest) (model.Booking, error) {
	return model.Booking{}, nil
}

func (b *bookingRepoStub) GetMyBookings(_ context.Context, _ uuid.UUID) ([]model.Booking, error) {
	return nil, nil
}

func (b *bookingRepoStub) GetAllBookings(_ context.Context, _ dto.BookingFilter) ([]model.Booking, int, error) {
	return b.allResult, b.allTotal, b.allErr
}

func TestBookingService_CreateBooking_Success(t *testing.T) {
	repo := &bookingRepoStub{createResult: model.Booking{ID: uuid.New()}}
	svc := service.NewBookingService(repo)

	res, err := svc.CreateBooking(context.Background(), uuid.New(), dto.BookingRequest{SlotID: uuid.New()})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if res.ID == uuid.Nil {
		t.Fatalf("expected non-empty booking id")
	}
}

func TestBookingService_CreateBooking_Error(t *testing.T) {
	repo := &bookingRepoStub{createErr: errors.New("repo error")}
	svc := service.NewBookingService(repo)

	_, err := svc.CreateBooking(context.Background(), uuid.New(), dto.BookingRequest{SlotID: uuid.New()})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestBookingService_GetAllBookings_ReturnsTotal(t *testing.T) {
	repo := &bookingRepoStub{allResult: []model.Booking{{ID: uuid.New()}}, allTotal: 42}
	svc := service.NewBookingService(repo)

	bookings, total, err := svc.GetAllBookings(context.Background(), dto.BookingFilter{Page: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(bookings) != 1 {
		t.Fatalf("expected 1 booking, got %d", len(bookings))
	}
	if total != 42 {
		t.Fatalf("expected total 42, got %d", total)
	}
}
