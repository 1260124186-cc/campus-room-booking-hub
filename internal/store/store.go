package store

import (
	"context"

	"github.com/zhangchengcheng/campus-room-booking-hub/internal/domain"
)

type Catalog interface {
	ListRooms(ctx context.Context) []domain.Room
	GetRoom(ctx context.Context, id string) (domain.Room, error)
}

type Bookings interface {
	Reserve(ctx context.Context, booking domain.Booking) error
	HasConflict(ctx context.Context, booking domain.Booking) bool
	Save(ctx context.Context, booking domain.Booking) error
	GetBooking(ctx context.Context, id string) (domain.Booking, error)
	ListBookings(ctx context.Context, date string) []domain.Booking
	CheckIn(ctx context.Context, id string) (domain.Booking, error)
}
