package handler

import (
	"RoomBookingService/internal/dto"
	"RoomBookingService/internal/httpresp"
	"RoomBookingService/internal/logger"
	"RoomBookingService/internal/usecase/service"
	"encoding/json"
	"net/http"
)

type RoomHandler struct {
	roomRepo service.RoomService
}

func NewRoomHandler(roomRepo service.RoomService) *RoomHandler {
	return &RoomHandler{
		roomRepo: roomRepo,
	}
}

func (h *RoomHandler) RegisterRoomRoutes(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())

	if r.Method != http.MethodPost {
		log.Warn("register room: method not allowed", "method", r.Method)
		httpresp.WriteError(w, http.StatusMethodNotAllowed, "INVALID_REQUEST", "method not allowed")
		return
	}

	var req dto.CreateRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warn("register room: invalid body", "error", err)
		httpresp.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	room, err := h.roomRepo.CreateRoom(r.Context(), req)
	if err != nil {
		log.Error("register room: service error", "error", err)
		httpresp.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}
	httpresp.WriteJSON(w, http.StatusCreated, map[string]any{"room": room})
}

func (h *RoomHandler) GetAllRooms(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())

	if r.Method != http.MethodGet {
		log.Warn("get all rooms: method not allowed", "method", r.Method)
		httpresp.WriteError(w, http.StatusMethodNotAllowed, "INVALID_REQUEST", "method not allowed")
		return
	}
	result, err := h.roomRepo.GetAllRooms(r.Context())
	if err != nil {
		log.Error("get all rooms: service error", "error", err)
		httpresp.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}
	httpresp.WriteJSON(w, http.StatusOK, map[string]any{"rooms": result})
}
