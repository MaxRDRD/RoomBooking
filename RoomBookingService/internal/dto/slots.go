package dto

import (
	"time"

	"github.com/google/uuid"
)

type SlotsRequest struct {
	RoomID uuid.UUID `json:"room_id" validate:"required"`
	Date   time.Time `json:"date" validate:"required"`
}
