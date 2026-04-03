package postgres

import (
	"RoomBookingService/internal/dto"
	"RoomBookingService/internal/logger"
	"RoomBookingService/internal/model"
	myerrors "RoomBookingService/pkg/my_errors"
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type slotRepository struct {
	db *pgxpool.Pool
}

func NewSlotRepository(db *pgxpool.Pool) *slotRepository {
	return &slotRepository{db: db}
}

func (r *slotRepository) GetAvailableSlots(ctx context.Context, req dto.SlotsRequest) ([]model.Slot, error) {
	log := logger.FromContext(ctx)
	log.Info("fetching available slots", "roomID", req.RoomID, "date", req.Date)

	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead})
	if err != nil {
		log.Error("failed to begin tx", "error", err)
		return nil, err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	var roomExists bool
	err = tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM rooms WHERE id = $1)`, req.RoomID).Scan(&roomExists)
	if err != nil {
		log.Error("failed to check room existence", "error", err)
		return nil, err
	}
	if !roomExists {
		return nil, myerrors.ErrRoomNotFound
	}

	var daysOfWeek []int32
	var startTime string
	var endTime string
	err = tx.QueryRow(ctx, `
		SELECT days_of_week, to_char(start_time, 'HH24:MI'), to_char(end_time, 'HH24:MI')
		FROM schedules
		WHERE room_id = $1`, req.RoomID).Scan(&daysOfWeek, &startTime, &endTime)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			if err := tx.Commit(ctx); err != nil {
				return nil, err
			}
			return []model.Slot{}, nil
		}
		log.Error("failed to get schedule", "error", err)
		return nil, err
	}

	isoWeekday := int(req.Date.Weekday())
	if isoWeekday == 0 {
		isoWeekday = 7
	}
	dayAllowed := false
	for _, d := range daysOfWeek {
		if int(d) == isoWeekday {
			dayAllowed = true
			break
		}
	}
	if !dayAllowed {
		if err := tx.Commit(ctx); err != nil {
			return nil, err
		}
		return []model.Slot{}, nil
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO slots (room_id, start_time, end_time)
		SELECT $1, gs, gs + interval '30 minutes'
		FROM generate_series(
			$2::date::timestamptz + $3::time,
			$2::date::timestamptz + $4::time - interval '30 minutes',
			interval '30 minutes'
		) AS gs
		ON CONFLICT (room_id, start_time) DO NOTHING`,
		req.RoomID,
		req.Date,
		startTime,
		endTime,
	)
	if err != nil {
		log.Error("failed to generate slots", "error", err)
		return nil, err
	}

	query := `
		SELECT s.id, s.room_id, s.start_time, s.end_time
		FROM slots s
		LEFT JOIN bookings b ON b.slot_id = s.id AND b.status = 'active'
		WHERE s.room_id = $1
		  AND s.start_time >= $2::date
		  AND s.start_time < ($2::date + interval '1 day')
		  AND b.id IS NULL
		ORDER BY s.start_time`
	rows, err := tx.Query(ctx, query, req.RoomID, req.Date)
	if err != nil {
		log.Error("failed to query available slots", "error", err)
		return nil, err
	}
	defer rows.Close()

	var slots []model.Slot
	for rows.Next() {
		var slot model.Slot
		if err := rows.Scan(&slot.ID, &slot.RoomID, &slot.StartAt, &slot.EndAt); err != nil {
			log.Error("failed to scan slot", "error", err)
			return nil, err
		}
		slots = append(slots, slot)
	}

	if err := rows.Err(); err != nil {
		log.Error("rows iteration error", "error", err)
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		log.Error("failed to commit tx", "error", err)
		return nil, err
	}

	return slots, nil
}
