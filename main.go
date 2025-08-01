package main

import (
	"log"
	"os"

	"github.com/bariiss/flarecert/cmd"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file if it exists
	if err := godotenv.Load(); err != nil {
		// Don't fail if .env doesn't exist, just log it
		log.Printf("Warning: .env file not found: %v", err)
	}

	// Execute the root command
	if err := cmd.Execute(); err != nil {
		log.Printf("Error: %v", err)
		os.Exit(1)
	}
}
