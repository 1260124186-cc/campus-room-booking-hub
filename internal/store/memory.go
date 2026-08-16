package store

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/zhangchengcheng/campus-room-booking-hub/internal/domain"
)

type MemoryStore struct {
	mu       sync.RWMutex
	rooms    map[string]domain.Room
	bookings map[string]domain.Booking
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		rooms: map[string]domain.Room{
			"atlas-101": {
				ID: "atlas-101", Name: "Atlas 101", Building: "Atlas", Capacity: 12,
				Equipment: []string{"whiteboard", "projector"},
				Manager:   &domain.Contact{Name: "Mina Ito", Email: "mina@example.edu"},
			},
			"atlas-202": {
				ID: "atlas-202", Name: "Atlas 202", Building: "Atlas", Capacity: 6,
				Equipment: []string{"whiteboard"},
			},
			"forum-3": {
				ID: "forum-3", Name: "Forum 3", Building: "Forum", Capacity: 30,
				Equipment: []string{"projector", "video-wall", "whiteboard"},
				Manager:   &domain.Contact{Name: "Ken Sato", Email: "ken@example.edu"},
			},
		},
		bookings: make(map[string]domain.Booking),
	}
}

func (s *MemoryStore) ListRooms(ctx context.Context) []domain.Room {
	if err := contextErr(ctx); err != nil {
		return nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]domain.Room, 0, len(s.rooms))
	for _, room := range s.rooms {
		result = append(result, cloneRoom(room))
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result
}

func (s *MemoryStore) GetRoom(ctx context.Context, id string) (domain.Room, error) {
	if err := contextErr(ctx); err != nil {
		return domain.Room{}, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	room, ok := s.rooms[id]
	if !ok {
		return domain.Room{}, domain.ErrRoomNotFound
	}
	return cloneRoom(room), nil
}

func (s *MemoryStore) Reserve(ctx context.Context, booking domain.Booking) error {
	if err := contextErr(ctx); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := contextErr(ctx); err != nil {
		return err
	}
	for _, existing := range s.bookings {
		if existing.RoomID == booking.RoomID && existing.Date == booking.Date && overlaps(existing, booking) {
			return domain.ErrSlotTaken
		}
	}
	s.bookings[booking.ID] = cloneBooking(booking)
	return nil
}

func (s *MemoryStore) GetBooking(ctx context.Context, id string) (domain.Booking, error) {
	if err := contextErr(ctx); err != nil {
		return domain.Booking{}, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	booking, ok := s.bookings[id]
	if !ok {
		return domain.Booking{}, domain.ErrBookingNotFound
	}
	return cloneBooking(booking), nil
}

func (s *MemoryStore) ListBookings(ctx context.Context, date string) []domain.Booking {
	if err := contextErr(ctx); err != nil {
		return nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]domain.Booking, 0)
	for _, booking := range s.bookings {
		if booking.Date == date {
			result = append(result, cloneBooking(booking))
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].RoomID == result[j].RoomID {
			return result[i].StartMinute < result[j].StartMinute
		}
		return result[i].RoomID < result[j].RoomID
	})
	return result
}

func (s *MemoryStore) CheckIn(ctx context.Context, id string) (domain.Booking, error) {
	if err := contextErr(ctx); err != nil {
		return domain.Booking{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := contextErr(ctx); err != nil {
		return domain.Booking{}, err
	}
	booking, ok := s.bookings[id]
	if !ok {
		return domain.Booking{}, domain.ErrBookingNotFound
	}
	if booking.Status == "checked_in" {
		return domain.Booking{}, domain.ErrAlreadyChecked
	}
	booking.Status = "checked_in"
	s.bookings[id] = booking
	return cloneBooking(booking), nil
}

func (s *MemoryStore) SeedBooking(booking domain.Booking) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.bookings[booking.ID] = cloneBooking(booking)
}

func contextErr(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

func overlaps(a, b domain.Booking) bool {
	aEnd := a.StartMinute + a.DurationMinutes
	bEnd := b.StartMinute + b.DurationMinutes
	return a.StartMinute < bEnd && b.StartMinute < aEnd
}

func cloneRoom(room domain.Room) domain.Room {
	room.Equipment = append([]string(nil), room.Equipment...)
	manager := *room.Manager
	room.Manager = &manager
	return room
}

func cloneBooking(booking domain.Booking) domain.Booking {
	booking.Equipment = append([]string(nil), booking.Equipment...)
	return booking
}

func NewBookingID(now time.Time, sequence int) string {
	return fmt.Sprintf("bk-%s-%03d", now.UTC().Format("20060102"), sequence)
}
