package repository

import (
	"RoomBookingService/internal/model"
	"context"
)

type UserRepository interface {
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)
	RegisterWithPassword(ctx context.Context, user *model.User) error
}
