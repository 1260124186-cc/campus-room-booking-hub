package store

import (
	"context"
	"testing"
)

func TestListedRoomEquipmentDoesNotChangeCatalog(t *testing.T) {
	memory := NewMemoryStore()
	rooms := memory.ListRooms(context.Background())
	rooms[0].Equipment[0] = "portable-speaker"

	room, err := memory.GetRoom(context.Background(), rooms[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	if room.Equipment[0] == "portable-speaker" {
		t.Fatal("listed room equipment changed the catalog")
	}
}

func TestRetrievedRoomEquipmentDoesNotChangeCatalog(t *testing.T) {
	memory := NewMemoryStore()
	room, err := memory.GetRoom(context.Background(), "atlas-101")
	if err != nil {
		t.Fatal(err)
	}
	room.Equipment[0] = "portable-speaker"

	room, err = memory.GetRoom(context.Background(), "atlas-101")
	if err != nil {
		t.Fatal(err)
	}
	if room.Equipment[0] == "portable-speaker" {
		t.Fatal("retrieved room equipment changed the catalog")
	}
}
