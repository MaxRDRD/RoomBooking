package dto

type CreateRoomRequest struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
	Capacity    *int   `json:"capacity"`
}
