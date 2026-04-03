package myerrors

import "errors"

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrUserUnauthorized   = errors.New("user unauthorized")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidRole        = errors.New("invalid role")

	ErrRoomNotFound      = errors.New("room not found")
	ErrRoomAlreadyExists = errors.New("room already exists")

	ErrScheduleNotFound      = errors.New("schedule not found")
	ErrScheduleAlreadyExists = errors.New("schedule already exists")

	ErrSlotNotFound      = errors.New("slot not found")
	ErrSlotAlreadyBooked = errors.New("slot already booked")
	ErrBookingNotFound   = errors.New("booking not found")
	ErrForbidden         = errors.New("forbidden")
	ErrInvalidRequest    = errors.New("invalid request")
)
