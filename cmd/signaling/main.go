package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/javanhut/zero/signaling"
)

func main() {
	server := signaling.NewServer()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down signaling server...")
		os.Exit(0)
	}()

	addr := ":8080"
	log.Printf("Starting Zero signaling server on %s", addr)

	if err := server.Start(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
