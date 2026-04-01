package repository

import (
	"RoomBookingService/internal/model"
	"context"
)

type UserRepository interface {
	Register(ctx context.Context, user *model.User) error
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)
}
