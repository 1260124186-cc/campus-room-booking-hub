package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/zhangchengcheng/campus-room-booking-hub/internal/service"
	"github.com/zhangchengcheng/campus-room-booking-hub/internal/store"
	"github.com/zhangchengcheng/campus-room-booking-hub/internal/transport"
)

func main() {
	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = ":8080"
	}
	memory := store.NewMemoryStore()
	svc := service.New(memory, memory)
	server := &http.Server{
		Addr:              addr,
		Handler:           transport.NewHandler(svc),
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Printf("campus room booking hub listening on %s", addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
