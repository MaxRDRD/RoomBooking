package unit_test

import (
	"RoomBookingService/internal/dto"
	"RoomBookingService/internal/model"
	"RoomBookingService/internal/usecase/service"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

type slotRepoStub struct {
	result []model.Slot
	err    error
}

func (s *slotRepoStub) GetAvailableSlots(_ context.Context, _ dto.SlotsRequest) ([]model.Slot, error) {
	return s.result, s.err
}

func TestSlotService_GetAvailableSlots_Success(t *testing.T) {
	repo := &slotRepoStub{result: []model.Slot{{ID: uuid.New(), RoomID: uuid.New(), StartAt: time.Now().UTC(), EndAt: time.Now().UTC().Add(30 * time.Minute)}}}
	svc := service.NewSlotService(repo)

	slots, err := svc.GetAvailableSlots(context.Background(), dto.SlotsRequest{RoomID: uuid.New(), Date: time.Now().UTC()})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(slots) != 1 {
		t.Fatalf("expected 1 slot, got %d", len(slots))
	}
}

func TestSlotService_GetAvailableSlots_Error(t *testing.T) {
	repo := &slotRepoStub{err: errors.New("query failed")}
	svc := service.NewSlotService(repo)

	_, err := svc.GetAvailableSlots(context.Background(), dto.SlotsRequest{RoomID: uuid.New(), Date: time.Now().UTC()})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}
