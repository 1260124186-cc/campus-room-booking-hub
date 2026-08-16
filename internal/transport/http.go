package transport

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/zhangchengcheng/campus-room-booking-hub/internal/domain"
	"github.com/zhangchengcheng/campus-room-booking-hub/internal/service"
)

type Handler struct {
	service *service.Service
}

func NewHandler(svc *service.Service) http.Handler {
	h := &Handler{service: svc}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", h.health)
	mux.HandleFunc("GET /rooms", h.rooms)
	mux.HandleFunc("POST /bookings", h.createBooking)
	mux.HandleFunc("GET /bookings/{id}", h.getBooking)
	mux.HandleFunc("POST /bookings/{id}/check-in", h.checkIn)
	mux.HandleFunc("GET /reports/occupancy", h.occupancy)
	return mux
}

func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) rooms(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"rooms": h.service.ListRooms(r.Context())})
}

func (h *Handler) createBooking(w http.ResponseWriter, r *http.Request) {
	var request domain.BookingRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "request body must be valid JSON")
		return
	}
	booking, err := h.service.CreateBooking(r.Context(), request)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, booking)
}

func (h *Handler) getBooking(w http.ResponseWriter, r *http.Request) {
	booking, err := h.service.GetBooking(r.Context(), r.PathValue("id"))
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, booking)
}

func (h *Handler) checkIn(w http.ResponseWriter, r *http.Request) {
	booking, err := h.service.CheckIn(r.Context(), r.PathValue("id"))
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, booking)
}

func (h *Handler) occupancy(w http.ResponseWriter, r *http.Request) {
	date := r.URL.Query().Get("date")
	if _, err := time.Parse("2006-01-02", date); err != nil {
		writeError(w, http.StatusBadRequest, "date must use YYYY-MM-DD")
		return
	}
	report, err := h.service.Occupancy(r.Context(), date)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"date": date, "rooms": report})
}

func (h *Handler) writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrRoomNotFound), errors.Is(err, domain.ErrBookingNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, domain.ErrSlotTaken):
		writeError(w, http.StatusConflict, err.Error())
	case errors.Is(err, domain.ErrAlreadyChecked):
		writeError(w, http.StatusConflict, err.Error())
	case errors.Is(err, domain.ErrInvalidBooking):
		writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		writeError(w, http.StatusRequestTimeout, "request canceled")
	default:
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}

func decodeJSON(r *http.Request, target any) error {
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(target); err != nil {
		return err
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
