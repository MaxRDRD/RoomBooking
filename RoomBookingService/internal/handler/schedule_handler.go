package handler

import (
	"RoomBookingService/internal/dto"
	"RoomBookingService/internal/httpresp"
	"RoomBookingService/internal/logger"
	"RoomBookingService/internal/usecase/service"
	myerrors "RoomBookingService/pkg/my_errors"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator"
	"github.com/google/uuid"
)

type ScheduleHandler struct {
	scheduleService service.ScheduleService
	validate        *validator.Validate
}

func NewScheduleHandler(scheduleService service.ScheduleService, validate *validator.Validate) *ScheduleHandler {
	return &ScheduleHandler{
		scheduleService: scheduleService,
		validate:        validate,
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

	roomID := chi.URLParam(r, "roomId")
	roomUUID, err := uuid.Parse(roomID)
	if err != nil {
		log.Warn("create schedule: invalid roomId", "roomId", roomID, "error", err)
		httpresp.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid roomId")
		return
	}
	req.RoomID = roomUUID

	if err := h.validate.Struct(req); err != nil {
		log.Warn("create schedule: validation failed", "error", err)
		httpresp.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid schedule payload")
		return
	}

	result, err := h.scheduleService.CreateSchedule(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, myerrors.ErrRoomNotFound):
			httpresp.WriteError(w, http.StatusNotFound, "ROOM_NOT_FOUND", "room not found")
		case errors.Is(err, myerrors.ErrScheduleAlreadyExists):
			httpresp.WriteError(w, http.StatusConflict, "SCHEDULE_EXISTS", "schedule for this room already exists and cannot be changed")
		case errors.Is(err, myerrors.ErrInvalidRequest):
			httpresp.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid schedule payload")
		default:
			log.Error("create schedule: service error", "error", err)
			httpresp.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		}
		return
	}
	httpresp.WriteJSON(w, http.StatusCreated, map[string]any{"schedule": result})
}
