# Campus Room Booking Hub

Campus Room Booking Hub is a self-contained Go HTTP service for coordinating meeting rooms on a campus. Students, faculty, and event coordinators can browse rooms, reserve a time window with capacity and equipment constraints, check in to a reservation, and inspect daily occupancy.

## Normal Features

- Room directory: list rooms, capacity, equipment, building, and the responsible contact.
- Reservation flow: validate room, date, time window, attendee details, seat count, equipment, and overlap before creating a reservation.
- Check-in flow: transition a reservation from `reserved` to `checked_in` exactly once.
- Occupancy report: summarize reservation count, reserved seats, and checked-in seats for a selected day.

## API

Start the service with `go run ./cmd/server`. It listens on `:8080` by default and respects the `ADDR` environment variable.

```sh
go build ./...
go test ./...
go run ./cmd/server
```

Useful requests:

```sh
curl http://127.0.0.1:8080/healthz
curl http://127.0.0.1:8080/rooms
curl 'http://127.0.0.1:8080/reports/occupancy?date=2026-09-10'
```

Create a booking with `POST /bookings`, then use `POST /bookings/{id}/check-in` to check in. The service has no external database or network dependency; all state is held in memory for the running process.
