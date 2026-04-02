package postgres

import (
	"RoomBookingService/internal/dto"
	"RoomBookingService/internal/logger"
	"RoomBookingService/internal/model"
	"context"

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

	query := `
		SELECT s.id, s.room_id, s.start_time, s.end_time
		FROM slots s
		LEFT JOIN bookings b ON b.slot_id = s.id AND b.status = 'active'
		WHERE s.room_id = $1
		  AND s.start_time >= $2::date
		  AND s.start_time < ($2::date + interval '1 day')
		  AND b.id IS NULL
		ORDER BY s.start_time`
	rows, err := r.db.Query(ctx, query, req.RoomID, req.Date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var slots []model.Slot
	for rows.Next() {
		var slot model.Slot
		if err := rows.Scan(&slot.ID, &slot.RoomID, &slot.StartAt, &slot.EndAt); err != nil {
			return nil, err
		}
		slots = append(slots, slot)
	}
	return slots, nil
}
