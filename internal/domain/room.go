package domain

import "time"

type Contact struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type Room struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Building  string   `json:"building"`
	Capacity  int      `json:"capacity"`
	Equipment []string `json:"equipment"`
	Manager   *Contact `json:"manager,omitempty"`
}

type BookingRequest struct {
	RoomID          string   `json:"room_id"`
	Attendee        string   `json:"attendee"`
	Email           string   `json:"email"`
	Date            string   `json:"date"`
	StartMinute     int      `json:"start_minute"`
	DurationMinutes int      `json:"duration_minutes"`
	Seats           int      `json:"seats"`
	Equipment       []string `json:"equipment"`
}

type Booking struct {
	ID              string    `json:"id"`
	RoomID          string    `json:"room_id"`
	Attendee        string    `json:"attendee"`
	Email           string    `json:"email"`
	Date            string    `json:"date"`
	StartMinute     int       `json:"start_minute"`
	DurationMinutes int       `json:"duration_minutes"`
	Seats           int       `json:"seats"`
	Equipment       []string  `json:"equipment"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
}

func (r Room) HasEquipment(required []string) bool {
	available := make(map[string]struct{}, len(r.Equipment))
	for _, item := range r.Equipment {
		available[item] = struct{}{}
	}
	for _, item := range required {
		if _, ok := available[item]; !ok {
			return false
		}
	}
	return true
}

func (r BookingRequest) Validate(room Room) error {
	if r.RoomID == "" || r.RoomID != room.ID || r.Attendee == "" || r.Email == "" {
		return ErrInvalidBooking
	}
	if r.Date == "" {
		return ErrInvalidBooking
	}
	if _, err := time.Parse("2006-01-02", r.Date); err != nil {
		return ErrInvalidBooking
	}
	if r.StartMinute < 8*60 || r.StartMinute >= 20*60 || r.DurationMinutes <= 0 || r.DurationMinutes > 240 {
		return ErrInvalidBooking
	}
	if r.StartMinute+r.DurationMinutes > 20*60 || r.Seats <= 0 || r.Seats > room.Capacity {
		return ErrInvalidBooking
	}
	if !room.HasEquipment(r.Equipment) {
		return ErrInvalidBooking
	}
	return nil
}
