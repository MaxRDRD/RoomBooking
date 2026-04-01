package postgres

import (
	"RoomBookingService/internal/model"
	"RoomBookingService/internal/repository"
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type userRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) repository.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Register(ctx context.Context, user *model.User) error {
	query := `INSERT INTO users (
	email, role, created_at
	) 
	VALUES ($1, $2, NOW())
	RETURNING id`
	err := r.db.QueryRow(ctx, query, user.Email, user.Role).Scan(&user.ID)
	return err
}

func (r *userRepository) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	query := `SELECT id, email, role, created_at FROM users WHERE email = $1`
	row := r.db.QueryRow(ctx, query, email)
	var user model.User
	err := row.Scan(&user.ID, &user.Email, &user.Role, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
