package main

import (
	"log"
	"os"

	"github.com/bariiss/flarecert/cmd"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file if it exists
	// Silently ignore if .env file doesn't exist (use system env vars)
	godotenv.Load()

	// Execute the root command
	if err := cmd.Execute(); err != nil {
		log.Printf("Error: %v", err)
		os.Exit(1)
	}
}
