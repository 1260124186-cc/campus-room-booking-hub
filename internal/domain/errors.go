package domain

import "errors"

var (
	ErrRoomNotFound    = errors.New("room not found")
	ErrBookingNotFound = errors.New("booking not found")
	ErrSlotTaken       = errors.New("room slot is already booked")
	ErrInvalidBooking  = errors.New("invalid booking")
	ErrAlreadyChecked  = errors.New("booking is already checked in")
)
