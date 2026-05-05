package repository

import (
	"RoomBookingService/internal/dto"
	"RoomBookingService/internal/model"
	"context"
)

type RoomRepository interface {
	CreateRoom(ctx context.Context, req dto.CreateRoomRequest) (model.Room, error)
	GetAllRooms(ctx context.Context) ([]model.Room, error) // доступно всем
}
