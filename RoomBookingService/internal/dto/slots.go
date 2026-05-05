package dto

import "github.com/google/uuid"

type SlotsRequest struct {
	RoomID uuid.UUID `json:"roomId" binding:"required"`
	Date   string    `json:"date" binding:"required"`
}
