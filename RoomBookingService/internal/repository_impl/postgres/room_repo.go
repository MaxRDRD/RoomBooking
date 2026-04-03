package postgres

import (
	"RoomBookingService/internal/model"
	myerrors "RoomBookingService/pkg/my_errors"
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type roomRepository struct {
	db *pgxpool.Pool
}

func NewRoomRepository(db *pgxpool.Pool) *roomRepository {
	return &roomRepository{db: db}
}

func (r *roomRepository) CreateRoom(ctx context.Context, room *model.Room) error {
	query := `INSERT INTO rooms (name, description, capacity, created_at) VALUES ($1, $2, $3, NOW()) RETURNING id`
	err := r.db.QueryRow(ctx, query, room.Name, room.Description, room.Capacity).Scan(&room.ID)
	return err
}

func (r *roomRepository) GetAllRooms(ctx context.Context) ([]*model.Room, error) {
	query := `SELECT id, name, description, capacity, created_at FROM rooms`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rooms []*model.Room
	for rows.Next() {
		var room model.Room
		err := rows.Scan(&room.ID, &room.Name, &room.Description, &room.Capacity, &room.CreatedAt)
		if err != nil {
			return nil, myerrors.ErrRoomNotFound
		}
		rooms = append(rooms, &room)
	}

	return rooms, nil
}
