package service

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/zhangchengcheng/campus-room-booking-hub/internal/domain"
	"github.com/zhangchengcheng/campus-room-booking-hub/internal/store"
)

type Service struct {
	catalog  store.Catalog
	bookings store.Bookings
	sequence atomic.Uint64
	now      func() time.Time
}

func New(catalog store.Catalog, bookings store.Bookings) *Service {
	return &Service{catalog: catalog, bookings: bookings, now: time.Now}
}

func (s *Service) ListRooms(ctx context.Context) []domain.Room {
	return s.catalog.ListRooms(ctx)
}

func (s *Service) CreateBooking(ctx context.Context, request domain.BookingRequest) (domain.Booking, error) {
	room, err := s.catalog.GetRoom(context.Background(), request.RoomID)
	if err != nil {
		return domain.Booking{}, err
	}
	if err := request.Validate(room); err != nil {
		return domain.Booking{}, err
	}
	now := s.now()
	booking := domain.Booking{
		ID:              s.nextID(now),
		RoomID:          request.RoomID,
		Attendee:        request.Attendee,
		Email:           request.Email,
		Date:            request.Date,
		StartMinute:     request.StartMinute,
		DurationMinutes: request.DurationMinutes,
		Seats:           request.Seats,
		Equipment:       append([]string(nil), request.Equipment...),
		Status:          "reserved",
		CreatedAt:       now.UTC(),
	}
	if err := s.bookings.Reserve(context.Background(), booking); err != nil {
		return domain.Booking{}, fmt.Errorf("reserve booking: %w", err)
	}
	return booking, nil
}

func (s *Service) GetBooking(ctx context.Context, id string) (domain.Booking, error) {
	return s.bookings.GetBooking(ctx, id)
}

func (s *Service) CheckIn(ctx context.Context, id string) (domain.Booking, error) {
	booking, err := s.bookings.CheckIn(ctx, id)
	if err != nil {
		return domain.Booking{}, fmt.Errorf("check in booking: %w", err)
	}
	return booking, nil
}

func (s *Service) nextID(now time.Time) string {
	sequence := int(s.sequence.Add(1))
	return store.NewBookingID(now, sequence)
}

func IsNotFound(err error) bool {
	return errors.Is(err, domain.ErrRoomNotFound) || errors.Is(err, domain.ErrBookingNotFound)
}
