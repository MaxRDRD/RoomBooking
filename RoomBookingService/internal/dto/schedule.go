package dto

import (
	"time"

	"github.com/google/uuid"
)

type CreateScheduleRequest struct {
	RoomID     uuid.UUID `json:"roomId" binding:"required"`
	DaysOfWeek []int     `json:"daysOfWeek" binding:"required"`
	StartTime  time.Time `json:"startTime" binding:"required"`
	EndTime    time.Time `json:"endTime" binding:"required"`
}
