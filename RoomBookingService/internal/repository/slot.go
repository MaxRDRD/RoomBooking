package repository

import (
	"RoomBookingService/internal/dto"
	"RoomBookingService/internal/model"
	"context"
)

type SlotRepository interface {
	GetAvailableSlots(ctx context.Context, req dto.SlotsRequest) ([]model.Slot, error)
}
