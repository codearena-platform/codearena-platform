package main

import (
	"log"
	"os"

	"github.com/codearena-platform/codearena-platform/internal/gateway"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	iamURL := os.Getenv("IAM_URL")
	if iamURL == "" {
		iamURL = "http://iam:7000"
	}

	compURL := os.Getenv("COMPETITION_URL")
	if compURL == "" {
		compURL = "http://competition:7001"
	}

	realtimeURL := os.Getenv("REALTIME_URL")
	if realtimeURL == "" {
		realtimeURL = "http://realtime:8081"
	}

	cfg := gateway.Config{
		Port:           port,
		IAMURL:         iamURL,
		CompetitionURL: compURL,
		RealtimeURL:    realtimeURL,
	}

	if err := gateway.Start(cfg); err != nil {
		log.Fatalf("gateway stopped: %v", err)
	}
}
