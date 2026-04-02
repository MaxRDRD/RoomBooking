package httpresp

import (
	"RoomBookingService/internal/model"
	"encoding/json"
	"net/http"
)

func WriteError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(model.JsonResponse{
		Error: model.ErrorDetail{
			Code:    code,
			Message: message,
		},
	})
}

func WriteJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
