package repository

import "RoomBookingService/internal/model"

type UserRepository interface {
	Register(email, password, role string) error
	Login(email, password string) (string, error)
	GetUserByEmail(email string) (*model.User, error)
}
