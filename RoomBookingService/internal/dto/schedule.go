package dto

import "github.com/google/uuid"

type CreateScheduleRequest struct {
	RoomID     uuid.UUID `json:"-" validate:"required"`
	DaysOfWeek []int     `json:"daysOfWeek" validate:"required,min=1,dive,min=1,max=7"`
	StartTime  string    `json:"startTime" validate:"required"`
	EndTime    string    `json:"endTime" validate:"required"`
}
