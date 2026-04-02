package postgres

import (
	"RoomBookingService/internal/model"
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type scheduleRepo struct {
	db *pgxpool.Pool
}

func NewScheduleRepository(db *pgxpool.Pool) *scheduleRepo {
	return &scheduleRepo{db: db}
}

func (r *scheduleRepo) CreateSchedule(ctx context.Context, schedule *model.Schedule) error {
	query := `INSERT INTO schedules (room_id, days_of_week, start_time, end_time, created_at) VALUES ($1, $2, $3, $4, NOW()) RETURNING id`
	err := r.db.QueryRow(ctx, query, schedule.RoomID, schedule.DaysOfWeek, schedule.StartTime, schedule.EndTime).Scan(&schedule.ID)
	return err
}
