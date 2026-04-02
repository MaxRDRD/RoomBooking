package handler

import (
	"RoomBookingService/internal/auth"
	"RoomBookingService/internal/dto"
	"RoomBookingService/internal/httpresp"
	"RoomBookingService/internal/logger"
	"RoomBookingService/internal/usecase/service"
	myerrors "RoomBookingService/pkg/my_errors"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type BookingHandler struct {
	bookingService service.BookingService
}

func NewBookingHandler(bookingService service.BookingService) *BookingHandler {
	return &BookingHandler{bookingService: bookingService}
}

func (h *BookingHandler) CreateBooking(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())

	var req dto.BookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpresp.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	if req.SlotID == uuid.Nil {
		httpresp.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "slotId is required")
		return
	}

	userID, err := auth.UserIDFromContext(r.Context())
	if err != nil {
		httpresp.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}

	booking, err := h.bookingService.CreateBooking(r.Context(), userID, req)
	if err != nil {
		switch {
		case errors.Is(err, myerrors.ErrSlotNotFound):
			httpresp.WriteError(w, http.StatusNotFound, "SLOT_NOT_FOUND", "slot not found")
		case errors.Is(err, myerrors.ErrSlotAlreadyBooked):
			httpresp.WriteError(w, http.StatusConflict, "SLOT_ALREADY_BOOKED", "slot is already booked")
		case errors.Is(err, myerrors.ErrInvalidRequest):
			httpresp.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "cannot create booking for past slot")
		default:
			log.Error("create booking: service error", "error", err)
			httpresp.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		}
		return
	}

	httpresp.WriteJSON(w, http.StatusCreated, map[string]any{"booking": booking})
}

func (h *BookingHandler) CancelBooking(w http.ResponseWriter, r *http.Request) {
	bookingIDStr := chi.URLParam(r, "bookingId")
	bookingID, err := uuid.Parse(bookingIDStr)
	if err != nil {
		httpresp.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid bookingId")
		return
	}

	userID, err := auth.UserIDFromContext(r.Context())
	if err != nil {
		httpresp.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}

	booking, err := h.bookingService.CancelBooking(r.Context(), userID, dto.CancelBookingRequest{BookingID: bookingID})
	if err != nil {
		switch {
		case errors.Is(err, myerrors.ErrBookingNotFound):
			httpresp.WriteError(w, http.StatusNotFound, "BOOKING_NOT_FOUND", "booking not found")
		case errors.Is(err, myerrors.ErrForbidden):
			httpresp.WriteError(w, http.StatusForbidden, "FORBIDDEN", "cannot cancel another user's booking")
		default:
			httpresp.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		}
		return
	}

	httpresp.WriteJSON(w, http.StatusOK, map[string]any{"booking": booking})
}

func (h *BookingHandler) GetMyBookings(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.UserIDFromContext(r.Context())
	if err != nil {
		httpresp.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}

	bookings, err := h.bookingService.GetMyBookings(r.Context(), userID, dto.BookingFilter{})
	if err != nil {
		httpresp.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}

	httpresp.WriteJSON(w, http.StatusOK, map[string]any{"bookings": bookings})
}

func (h *BookingHandler) GetAllBookings(w http.ResponseWriter, r *http.Request) {
	page := 1
	pageSize := 20

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		p, err := strconv.Atoi(pageStr)
		if err != nil || p < 1 {
			httpresp.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid page")
			return
		}
		page = p
	}

	if pageSizeStr := r.URL.Query().Get("pageSize"); pageSizeStr != "" {
		ps, err := strconv.Atoi(pageSizeStr)
		if err != nil || ps < 1 || ps > 100 {
			httpresp.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid pageSize")
			return
		}
		pageSize = ps
	}

	filter := dto.BookingFilter{Page: page, PageSize: pageSize}
	bookings, total, err := h.bookingService.GetAllBookings(r.Context(), filter)
	if err != nil {
		httpresp.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}

	httpresp.WriteJSON(w, http.StatusOK, map[string]any{
		"bookings": bookings,
		"pagination": map[string]any{
			"page":     page,
			"pageSize": pageSize,
			"total":    total,
		},
	})
}
