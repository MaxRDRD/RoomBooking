package dto

import (
	"time"

	"github.com/google/uuid"
)

type BookingRequest struct {
	SlotID               uuid.UUID `json:"slotId"`
	CreateConferenceLink bool      `json:"createConferenceLink"`
}

type CancelBookingRequest struct {
	BookingID uuid.UUID `json:"bookingId"`
}

type BookingFilter struct {
	UserID   *uuid.UUID `json:"user_id,omitempty"`
	RoomID   *uuid.UUID `json:"room_id,omitempty"`
	From     *time.Time `json:"from"`
	To       *time.Time `json:"to"`
	Page     int        `json:"page"`
	PageSize int        `json:"pageSize"`
}
