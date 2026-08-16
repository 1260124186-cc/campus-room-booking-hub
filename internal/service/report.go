package service

import (
	"context"
	"sort"
)

type Occupancy struct {
	RoomID         string `json:"room_id"`
	RoomName       string `json:"room_name"`
	Capacity       int    `json:"capacity"`
	ReservedSeats  int    `json:"reserved_seats"`
	CheckedInSeats int    `json:"checked_in_seats"`
	BookingCount   int    `json:"booking_count"`
}

func (s *Service) Occupancy(ctx context.Context, date string) ([]Occupancy, error) {
	rooms := s.catalog.ListRooms(ctx)
	bookings := s.bookings.ListBookings(ctx, date)
	byRoom := make(map[string]*Occupancy, len(rooms))
	for _, room := range rooms {
		room := room
		byRoom[room.ID] = &Occupancy{RoomID: room.ID, RoomName: room.Name, Capacity: room.Capacity}
	}
	for _, booking := range bookings {
		item, ok := byRoom[booking.RoomID]
		if !ok {
			continue
		}
		item.BookingCount++
		item.ReservedSeats += booking.Seats
		if booking.Status == "checked_in" {
			item.CheckedInSeats += booking.Seats
		}
	}
	result := make([]Occupancy, 0, len(byRoom))
	for _, item := range byRoom {
		result = append(result, *item)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].RoomID < result[j].RoomID })
	return result, nil
}
