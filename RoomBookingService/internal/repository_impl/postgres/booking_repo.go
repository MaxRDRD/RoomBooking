package postgres

import (
	"RoomBookingService/internal/dto"
	"RoomBookingService/internal/logger"
	"RoomBookingService/internal/model"
	myerrors "RoomBookingService/pkg/my_errors"
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type bookingRepository struct {
	db *pgxpool.Pool
}

func NewBookingRepository(db *pgxpool.Pool) *bookingRepository {
	return &bookingRepository{db: db}
}

func (r *bookingRepository) CreateBooking(ctx context.Context, userID uuid.UUID, req dto.BookingRequest) (model.Booking, error) {
	log := logger.FromContext(ctx)

	query := `
		INSERT INTO bookings (user_id, slot_id, status, conference_link, created_at)
		SELECT $1, s.id, 'active', $2, NOW()
		FROM slots s
		WHERE s.id = $3
		  AND s.start_time >= NOW()
		RETURNING id, user_id, slot_id, status, conference_link, created_at`

	var booking model.Booking
	conferenceLink := any(nil)

	err := r.db.QueryRow(ctx, query, userID, conferenceLink, req.SlotID).Scan(
		&booking.ID,
		&booking.UserID,
		&booking.SlotID,
		&booking.Status,
		&booking.ConferenceLink,
		&booking.CreatedAt,
	)
	if err == nil {
		return booking, nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		log.Debug("slot already booked", "slot_id", req.SlotID)
		return model.Booking{}, myerrors.ErrSlotAlreadyBooked
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		log.Error("failed to create booking", "error", err)
		return model.Booking{}, err
	}

	var exists bool
	err = r.db.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM slots WHERE id = $1)`, req.SlotID).Scan(&exists)
	if err != nil {
		log.Error("failed to check slot existence", "error", err)
		return model.Booking{}, err
	}
	if !exists {
		return model.Booking{}, myerrors.ErrSlotNotFound
	}

	return model.Booking{}, myerrors.ErrInvalidRequest
}

func (r *bookingRepository) CancelBooking(ctx context.Context, userID uuid.UUID, req dto.CancelBookingRequest) (model.Booking, error) {
	log := logger.FromContext(ctx)

	var booking model.Booking
	err := r.db.QueryRow(ctx, `
		SELECT id, user_id, slot_id, status, conference_link, created_at
		FROM bookings
		WHERE id = $1`, req.BookingID).Scan(
		&booking.ID,
		&booking.UserID,
		&booking.SlotID,
		&booking.Status,
		&booking.ConferenceLink,
		&booking.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Error("booking not found", "error", err)
			return model.Booking{}, myerrors.ErrBookingNotFound
		}
		log.Error("failed to fetch booking", "error", err)
		return model.Booking{}, err
	}

	if booking.UserID != userID {
		log.Error("forbidden", "error", err)
		return model.Booking{}, myerrors.ErrForbidden
	}

	if booking.Status == "cancelled" {
		return booking, nil
	}

	err = r.db.QueryRow(ctx, `
		UPDATE bookings
		SET status = 'cancelled'
		WHERE id = $1
		RETURNING id, user_id, slot_id, status, conference_link, created_at`, req.BookingID).Scan(
		&booking.ID,
		&booking.UserID,
		&booking.SlotID,
		&booking.Status,
		&booking.ConferenceLink,
		&booking.CreatedAt,
	)
	if err != nil {
		log.Error("failed to cancel booking", "error", err)
		return model.Booking{}, err
	}

	return booking, nil
}

func (r *bookingRepository) GetMyBookings(ctx context.Context, userID uuid.UUID, _ dto.BookingFilter) ([]model.Booking, error) {
	log := logger.FromContext(ctx)
	query := `
		SELECT b.id, b.user_id, b.slot_id, b.status, b.conference_link, b.created_at
		FROM bookings b
		JOIN slots s ON s.id = b.slot_id
		WHERE b.user_id = $1
		  AND s.start_time >= NOW()
		ORDER BY s.start_time`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		log.Error("failed to get my bookings", "error", err)
		return nil, err
	}
	defer rows.Close()

	bookings := make([]model.Booking, 0)
	for rows.Next() {
		var b model.Booking
		if err := rows.Scan(&b.ID, &b.UserID, &b.SlotID, &b.Status, &b.ConferenceLink, &b.CreatedAt); err != nil {
			log.Error("failed to scan booking", "error", err)
			return nil, err
		}
		bookings = append(bookings, b)
	}

	return bookings, nil
}

func (r *bookingRepository) GetAllBookings(ctx context.Context, filter dto.BookingFilter) ([]model.Booking, int, error) {
	log := logger.FromContext(ctx)
	page := filter.Page
	if page <= 0 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	offset := (page - 1) * pageSize

	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead})
	if err != nil {
		log.Error("failed to begin transaction", "error", err)
		return nil, 0, err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	var total int
	err = tx.QueryRow(ctx, `SELECT COUNT(*) FROM bookings`).Scan(&total)
	if err != nil {
		log.Error("failed to count bookings", "error", err)
		return nil, 0, err
	}

	query := `
		SELECT id, user_id, slot_id, status, conference_link, created_at
		FROM bookings
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	rows, err := tx.Query(ctx, query, pageSize, offset)
	if err != nil {
		log.Error("failed to get all bookings", "error", err)
		return nil, 0, err
	}
	defer rows.Close()

	bookings := make([]model.Booking, 0)
	for rows.Next() {
		var b model.Booking
		if err := rows.Scan(&b.ID, &b.UserID, &b.SlotID, &b.Status, &b.ConferenceLink, &b.CreatedAt); err != nil {
			log.Error("failed to scan booking", "error", err)
			return nil, 0, err
		}
		bookings = append(bookings, b)
	}

	if err := rows.Err(); err != nil {
		log.Error("rows iteration error", "error", err)
		return nil, 0, err
	}

	if err := tx.Commit(ctx); err != nil {
		log.Error("failed to commit transaction", "error", err)
		return nil, 0, err
	}

	return bookings, total, nil
}
