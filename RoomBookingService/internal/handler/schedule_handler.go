package handler

import (
	"RoomBookingService/internal/dto"
	"RoomBookingService/internal/httpresp"
	"RoomBookingService/internal/logger"
	"RoomBookingService/internal/usecase/service"
	"encoding/json"
	"net/http"
)

type ScheduleHandler struct {
	scheduleService service.ScheduleService
}

func NewScheduleHandler(scheduleService service.ScheduleService) *ScheduleHandler {
	return &ScheduleHandler{
		scheduleService: scheduleService,
	}
}

func (h *ScheduleHandler) CreateSchedule(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())

	if r.Method != http.MethodPost {
		log.Warn("create schedule: method not allowed", "method", r.Method)
		httpresp.WriteError(w, http.StatusMethodNotAllowed, "INVALID_REQUEST", "method not allowed")
		return
	}

	var req dto.CreateScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warn("create schedule: invalid body", "error", err)
		httpresp.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	result, err := h.scheduleService.CreateSchedule(r.Context(), req)
	if err != nil {
		log.Error("create schedule: service error", "error", err)
		httpresp.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}
	httpresp.WriteJSON(w, http.StatusCreated, map[string]any{"schedule": result})
}
