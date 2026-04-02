package repository

import (
	"RoomBookingService/internal/model"
	"context"
)

type ScheduleRepository interface {
	CreateSchedule(ctx context.Context, schedule *model.Schedule) error
}
