package dto

import (
	"time"

	"github.com/google/uuid"
)

type SlotsRequest struct {
	RoomID uuid.UUID `json:"roomId" binding:"required"`
	Date   time.Time `json:"date" binding:"required"`
}
