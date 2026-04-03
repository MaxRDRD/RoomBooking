package handler

import (
	"RoomBookingService/internal/dto"
	"RoomBookingService/internal/httpresp"
	"RoomBookingService/internal/logger"
	"RoomBookingService/internal/usecase/service"
	myerrors "RoomBookingService/pkg/my_errors"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type SlotHandler struct {
	slotService service.SlotService
}

func NewSlotHandler(slotService service.SlotService) *SlotHandler {
	return &SlotHandler{slotService: slotService}
}

func (h *SlotHandler) GetAvailableSlots(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())

	if r.Method != http.MethodGet {
		log.Warn("get available slots: method not allowed", "method", r.Method)
		httpresp.WriteError(w, http.StatusMethodNotAllowed, "INVALID_REQUEST", "method not allowed")
		return
	}

	roomID := chi.URLParam(r, "roomId")
	date := r.URL.Query().Get("date")

	if date == "" || roomID == "" {
		log.Warn("get available slots: missing required parameters", "roomId", roomID, "date", date)
		httpresp.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "missing required parameters: roomId and date")
		return
	}

	roomUUID, err := uuid.Parse(roomID)
	if err != nil {
		log.Warn("get available slots: invalid roomId", "roomId", roomID, "error", err)
		httpresp.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid roomId")
		return
	}

	dateValue, err := time.Parse("2006-01-02", date)
	if err != nil {
		log.Warn("get available slots: invalid date format", "date", date, "error", err)
		httpresp.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid date format, expected YYYY-MM-DD")
		return
	}

	req := dto.SlotsRequest{
		RoomID: roomUUID,
		Date:   dateValue.UTC(),
	}

	slots, err := h.slotService.GetAvailableSlots(r.Context(), req)
	if err != nil {
		if errors.Is(err, myerrors.ErrRoomNotFound) {
			log.Warn("get available slots: room not found", "roomId", roomID)
			httpresp.WriteError(w, http.StatusNotFound, "ROOM_NOT_FOUND", "room not found")
			return
		}
		log.Error("get available slots: service error", "error", err)
		httpresp.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}

	httpresp.WriteJSON(w, http.StatusOK, map[string]any{"slots": slots})
}
