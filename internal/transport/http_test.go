package transport_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/zhangchengcheng/campus-room-booking-hub/internal/domain"
	"github.com/zhangchengcheng/campus-room-booking-hub/internal/service"
	"github.com/zhangchengcheng/campus-room-booking-hub/internal/store"
	"github.com/zhangchengcheng/campus-room-booking-hub/internal/transport"
)

func newHandler() http.Handler {
	memory := store.NewMemoryStore()
	svc := service.New(memory, memory)
	return transport.NewHandler(svc)
}

func perform(handler http.Handler, method, target string, body []byte) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, target, bytes.NewReader(body))
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	return recorder
}

func TestHealthAndRooms(t *testing.T) {
	handler := newHandler()

	response := perform(handler, http.MethodGet, "/healthz", nil)
	if response.Code != http.StatusOK {
		t.Fatalf("health status = %d", response.Code)
	}

	response = perform(handler, http.MethodGet, "/rooms", nil)
	var body struct {
		Rooms []domain.Room `json:"rooms"`
	}
	if err := json.NewDecoder(response.Result().Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if len(body.Rooms) != 3 {
		t.Fatalf("room count = %d, want 3", len(body.Rooms))
	}
}

func TestCreateCheckInAndReport(t *testing.T) {
	handler := newHandler()
	payload := domain.BookingRequest{
		RoomID: "atlas-101", Attendee: "Aiko", Email: "aiko@example.edu",
		Date: "2026-09-10", StartMinute: 600, DurationMinutes: 60,
		Seats: 4, Equipment: []string{"projector"},
	}
	body, _ := json.Marshal(payload)
	response := perform(handler, http.MethodPost, "/bookings", body)
	if response.Code != http.StatusCreated {
		t.Fatalf("create status = %d", response.Code)
	}
	var booking domain.Booking
	if err := json.NewDecoder(response.Result().Body).Decode(&booking); err != nil {
		t.Fatal(err)
	}

	response = perform(handler, http.MethodPost, "/bookings/"+booking.ID+"/check-in", nil)
	if response.Code != http.StatusOK {
		t.Fatalf("check-in status = %d", response.Code)
	}

	response = perform(handler, http.MethodGet, "/reports/occupancy?date=2026-09-10", nil)
	var report struct {
		Rooms []service.Occupancy `json:"rooms"`
	}
	if err := json.NewDecoder(response.Result().Body).Decode(&report); err != nil {
		t.Fatal(err)
	}
	var found service.Occupancy
	for _, room := range report.Rooms {
		if room.RoomID == "atlas-101" {
			found = room
		}
	}
	if found.CheckedInSeats != 4 || found.BookingCount != 1 {
		t.Fatalf("report = %+v", found)
	}
}

func TestRejectsInvalidBookingAndOverlap(t *testing.T) {
	handler := newHandler()
	payload := domain.BookingRequest{
		RoomID: "atlas-202", Attendee: "Noon", Email: "noon@example.edu",
		Date: "2026-09-10", StartMinute: 600, DurationMinutes: 60, Seats: 2,
	}
	body, _ := json.Marshal(payload)
	response := perform(handler, http.MethodPost, "/bookings", body)
	if response.Code != http.StatusCreated {
		t.Fatalf("first create status = %d", response.Code)
	}

	body, _ = json.Marshal(payload)
	response = perform(handler, http.MethodPost, "/bookings", body)
	if response.Code != http.StatusConflict {
		t.Fatalf("overlap status = %d", response.Code)
	}

	payload.StartMinute = 7 * 60
	body, _ = json.Marshal(payload)
	response = perform(handler, http.MethodPost, "/bookings", body)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("invalid status = %d", response.Code)
	}
}

func TestConcurrentReservationsHaveSingleWinner(t *testing.T) {
	memory := store.NewMemoryStore()
	svc := service.New(memory, memory)
	ctx := context.Background()
	request := domain.BookingRequest{
		RoomID: "forum-3", Attendee: "Group", Email: "group@example.edu",
		Date: "2026-09-11", StartMinute: 600, DurationMinutes: 120, Seats: 8,
	}
	const workers = 12
	var wg sync.WaitGroup
	var mu sync.Mutex
	successes := 0
	conflicts := 0
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := svc.CreateBooking(ctx, request)
			switch {
			case err == nil:
				mu.Lock()
				successes++
				mu.Unlock()
			case errors.Is(err, domain.ErrSlotTaken):
				mu.Lock()
				conflicts++
				mu.Unlock()
			default:
				t.Errorf("CreateBooking() error = %v, want ErrSlotTaken", err)
			}
		}()
	}
	wg.Wait()
	if successes != 1 {
		t.Fatalf("successful reservations = %d, want 1", successes)
	}
	if conflicts != workers-1 {
		t.Fatalf("conflicting reservations = %d, want %d", conflicts, workers-1)
	}
	if bookings := memory.ListBookings(ctx, request.Date); len(bookings) != 1 {
		t.Fatalf("persisted bookings = %d, want 1", len(bookings))
	}
}

func TestCanceledReservationDoesNotPersist(t *testing.T) {
	memory := store.NewMemoryStore()
	svc := service.New(memory, memory)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := svc.CreateBooking(ctx, domain.BookingRequest{
		RoomID: "atlas-101", Attendee: "Canceled", Email: "cancel@example.edu",
		Date: "2026-09-12", StartMinute: 600, DurationMinutes: 60, Seats: 1,
	})
	if err == nil {
		t.Fatal("expected cancellation error")
	}
	if bookings := memory.ListBookings(context.Background(), "2026-09-12"); len(bookings) != 0 {
		t.Fatalf("canceled request persisted %d booking(s)", len(bookings))
	}
}

func TestCheckInUnknownBookingReturnsNotFound(t *testing.T) {
	handler := newHandler()
	response := perform(handler, http.MethodPost, "/bookings/missing/check-in", nil)
	if response.Code != http.StatusNotFound {
		t.Fatalf("unknown check-in status = %d", response.Code)
	}
}

func TestContextDeadlineAtStoreBoundary(t *testing.T) {
	memory := store.NewMemoryStore()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := memory.Reserve(ctx, domain.Booking{ID: "bk-canceled", RoomID: "atlas-101", Date: "2026-09-13", StartMinute: 600, DurationMinutes: 30, Seats: 1})
	if err == nil {
		t.Fatal("expected canceled context error")
	}
}

func TestRequestTimeoutDoesNotHang(t *testing.T) {
	memory := store.NewMemoryStore()
	svc := service.New(memory, memory)
	ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
	defer cancel()
	time.Sleep(time.Millisecond)
	_, err := svc.CreateBooking(ctx, domain.BookingRequest{
		RoomID: "atlas-101", Attendee: "Late", Email: "late@example.edu",
		Date: "2026-09-14", StartMinute: 600, DurationMinutes: 30, Seats: 1,
	})
	if err == nil {
		t.Fatal("expected deadline error")
	}
}
