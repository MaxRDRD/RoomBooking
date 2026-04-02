package service

import (
	"RoomBookingService/internal/dto"
	"RoomBookingService/internal/logger"
	"RoomBookingService/internal/model"
	"RoomBookingService/internal/repository"
	"context"
)

type SlotService struct {
	slotRepo repository.SlotRepository
}

func NewSlotService(slotRepo repository.SlotRepository) *SlotService {
	return &SlotService{slotRepo: slotRepo}
}

func (s *SlotService) GetAvailableSlots(ctx context.Context, req dto.SlotsRequest) ([]model.Slot, error) {
	log := logger.FromContext(ctx)

	slots, err := s.slotRepo.GetAvailableSlots(ctx, req)
	if err != nil {
		log.Error("failed to get available slots", "error", err)
		return nil, err
	}
	return slots, nil
}
