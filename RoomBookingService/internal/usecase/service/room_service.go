package service

import (
	"RoomBookingService/internal/dto"
	"RoomBookingService/internal/logger"
	"RoomBookingService/internal/model"
	"RoomBookingService/internal/repository"
	"context"
)

type RoomService interface {
	CreateRoom(ctx context.Context, req dto.CreateRoomRequest) (*model.Room, error)
	GetAllRooms(ctx context.Context) ([]*model.Room, error)
}

type roomService struct {
	roomRepo repository.RoomRepository
}

func NewRoomService(roomRepo repository.RoomRepository) RoomService {
	return &roomService{
		roomRepo: roomRepo,
	}
}

func (s *roomService) CreateRoom(ctx context.Context, req dto.CreateRoomRequest) (*model.Room, error) {
	log := logger.FromContext(ctx)
	room := &model.Room{
		Name: req.Name,
	}
	if req.Capacity != nil {
		room.Capacity = *req.Capacity
	}
	err := s.roomRepo.CreateRoom(ctx, room)
	if err != nil {
		log.Error("failed to create room", "error", err)
		return nil, err
	}
	return room, nil
}

func (s *roomService) GetAllRooms(ctx context.Context) ([]*model.Room, error) {
	log := logger.FromContext(ctx)
	res, err := s.roomRepo.GetAllRooms(ctx)
	if err != nil {
		log.Error("failed to get all rooms", "error", err)
		return nil, err
	}
	return res, nil
}
