package unit_test

import (
	"RoomBookingService/internal/dto"
	"RoomBookingService/internal/model"
	"RoomBookingService/internal/usecase/service"
	myerrors "RoomBookingService/pkg/my_errors"
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

type scheduleRepoStub struct {
	createErr error
}

func (s *scheduleRepoStub) CreateSchedule(_ context.Context, _ *model.Schedule) error {
	return s.createErr
}

func TestScheduleService_CreateSchedule_Success(t *testing.T) {
	repo := &scheduleRepoStub{}
	svc := service.NewScheduleService(repo)

	res, err := svc.CreateSchedule(context.Background(), dto.CreateScheduleRequest{
		RoomID:     uuid.New(),
		DaysOfWeek: []int{1, 3, 5},
		StartTime:  "9:00",
		EndTime:    "18:30",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if res == nil {
		t.Fatalf("expected non-nil schedule")
	}
	if res.StartTime.Hour() != 9 || res.EndTime.Hour() != 18 {
		t.Fatalf("unexpected time conversion: start=%v end=%v", res.StartTime, res.EndTime)
	}
}

func TestScheduleService_CreateSchedule_InvalidTime(t *testing.T) {
	repo := &scheduleRepoStub{}
	svc := service.NewScheduleService(repo)

	_, err := svc.CreateSchedule(context.Background(), dto.CreateScheduleRequest{
		RoomID:     uuid.New(),
		DaysOfWeek: []int{1},
		StartTime:  "25:00",
		EndTime:    "18:00",
	})
	if !errors.Is(err, myerrors.ErrInvalidRequest) {
		t.Fatalf("expected ErrInvalidRequest, got %v", err)
	}
}

func TestScheduleService_CreateSchedule_MapsPgErrors(t *testing.T) {
	t.Run("schedule exists", func(t *testing.T) {
		repo := &scheduleRepoStub{createErr: &pgconn.PgError{Code: "23505"}}
		svc := service.NewScheduleService(repo)

		_, err := svc.CreateSchedule(context.Background(), dto.CreateScheduleRequest{
			RoomID:     uuid.New(),
			DaysOfWeek: []int{1},
			StartTime:  "9:00",
			EndTime:    "10:00",
		})
		if !errors.Is(err, myerrors.ErrScheduleAlreadyExists) {
			t.Fatalf("expected ErrScheduleAlreadyExists, got %v", err)
		}
	})

	t.Run("room not found", func(t *testing.T) {
		repo := &scheduleRepoStub{createErr: &pgconn.PgError{Code: "23503"}}
		svc := service.NewScheduleService(repo)

		_, err := svc.CreateSchedule(context.Background(), dto.CreateScheduleRequest{
			RoomID:     uuid.New(),
			DaysOfWeek: []int{1},
			StartTime:  "9:00",
			EndTime:    "10:00",
		})
		if !errors.Is(err, myerrors.ErrRoomNotFound) {
			t.Fatalf("expected ErrRoomNotFound, got %v", err)
		}
	})
}
