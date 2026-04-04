package postgres

import (
	"RoomBookingService/internal/model"
	"RoomBookingService/internal/repository"
	myerrors "RoomBookingService/pkg/my_errors"
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type userRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) repository.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	query := `SELECT id, email, COALESCE(password_hash, ''), role, created_at FROM users WHERE email = $1`
	row := r.db.QueryRow(ctx, query, email)
	var user model.User
	err := row.Scan(&user.ID, &user.Email, &user.Password, &user.Role, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, myerrors.ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) RegisterWithPassword(ctx context.Context, user *model.User) error {
	query := `INSERT INTO users (
	email, role, password_hash, created_at
	) 
	VALUES ($1, $2, $3, NOW())
	RETURNING id`
	err := r.db.QueryRow(ctx, query, user.Email, user.Role, user.Password).Scan(&user.ID)
	return err
}
