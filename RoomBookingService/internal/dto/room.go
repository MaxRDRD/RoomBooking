package dto

type CreateRoomRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Capacity    *int   `json:"capacity"`
}
