package model

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID    uuid.UUID
	Email string
	Role      string
	CreatedAt time.Time
}
