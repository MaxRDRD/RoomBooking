package service

import (
	"RoomBookingService/internal/dto"
	"RoomBookingService/internal/logger"
	"RoomBookingService/internal/model"
	"RoomBookingService/internal/repository"
	myerrors "RoomBookingService/pkg/my_errors"
	"context"
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
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

	start, err := parseTimeOfDay(req.StartTime)
	if err != nil {
		return nil, myerrors.ErrInvalidRequest
	}
	end, err := parseTimeOfDay(req.EndTime)
	if err != nil {
		return nil, myerrors.ErrInvalidRequest
	}

	start = time.Date(0, 1, 1, start.Hour(), start.Minute(), 0, 0, time.UTC)
	end = time.Date(0, 1, 1, end.Hour(), end.Minute(), 0, 0, time.UTC)
	if !start.Before(end) {
		return nil, myerrors.ErrInvalidRequest
	}

	schedule := &model.Schedule{
		RoomID:     req.RoomID,
		DaysOfWeek: req.DaysOfWeek,
		StartTime:  start,
		EndTime:    end,
	}

	err = s.scheduleRepo.CreateSchedule(ctx, schedule)
	if err != nil {
		log.Error("failed to create schedule", "error", err)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505":
				return nil, myerrors.ErrScheduleAlreadyExists
			case "23503":
				return nil, myerrors.ErrRoomNotFound
			}
		}
		return nil, err
	}
	return schedule, nil
}

func parseTimeOfDay(value string) (time.Time, error) {
	pattern := regexp.MustCompile(`^([01]?[0-9]|2[0-3]):[0-5][0-9]$`)
	if !pattern.MatchString(value) {
		return time.Time{}, myerrors.ErrInvalidRequest
	}

	parts := strings.Split(value, ":")
	if len(parts) != 2 {
		return time.Time{}, myerrors.ErrInvalidRequest
	}

	hour, err := strconv.Atoi(parts[0])
	if err != nil {
		return time.Time{}, myerrors.ErrInvalidRequest
	}
	minute, err := strconv.Atoi(parts[1])
	if err != nil {
		return time.Time{}, myerrors.ErrInvalidRequest
	}

	if hour < 0 || hour > 23 || minute < 0 || minute > 59 {
		return time.Time{}, myerrors.ErrInvalidRequest
	}

	return time.Date(0, 1, 1, hour, minute, 0, 0, time.UTC), nil
}
