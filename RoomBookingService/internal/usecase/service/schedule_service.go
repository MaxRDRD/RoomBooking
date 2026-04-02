package service

import (
	"RoomBookingService/internal/dto"
	"RoomBookingService/internal/logger"
	"RoomBookingService/internal/model"
	"RoomBookingService/internal/repository"
	myerrors "RoomBookingService/pkg/my_errors"
	"context"
)

type ScheduleService interface {
	CreateSchedule(ctx context.Context, req dto.CreateScheduleRequest) (*model.Schedule, error)
}

type scheduleService struct {
	scheduleRepo repository.ScheduleRepository
}

func NewScheduleService(scheduleRepo repository.ScheduleRepository) ScheduleService {
	return &scheduleService{
		scheduleRepo: scheduleRepo,
	}
}

func (s *scheduleService) CreateSchedule(ctx context.Context, req dto.CreateScheduleRequest) (*model.Schedule, error) {
	log := logger.FromContext(ctx)

	schedule := &model.Schedule{
		RoomID:     req.RoomID,
		DaysOfWeek: req.DaysOfWeek,
		StartTime:  req.StartTime,
		EndTime:    req.EndTime,
	}

	err := s.scheduleRepo.CreateSchedule(ctx, schedule)
	if err != nil {
		log.Error("failed to create schedule", "error", err)
		return nil, myerrors.ErrScheduleAlreadyExists
	}
	return schedule, nil
}
