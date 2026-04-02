package repository

import (
	"RoomBookingService/internal/model"
	"context"
)

type RoomRepository interface {
	CreateRoom(ctx context.Context, room *model.Room) error
	GetAllRooms(ctx context.Context) ([]*model.Room, error) // доступно всем
}
