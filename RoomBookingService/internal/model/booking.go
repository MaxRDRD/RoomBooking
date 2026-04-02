package model

import (
	"time"

	"github.com/google/uuid"
)

type Booking struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	SlotID    uuid.UUID
	StartTime time.Time
	Status    string
	ConferenceLink *string
	CreatedAt time.Time
}
