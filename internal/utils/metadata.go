package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// CertificateMetadata holds metadata about a certificate
type CertificateMetadata struct {
	Domain       string    `json:"domain"`
	Domains      []string  `json:"domains"`
	IsWildcard   bool      `json:"is_wildcard"`
	KeyType      string    `json:"key_type"`
	CreatedAt    time.Time `json:"created_at"`
	ExpiresAt    time.Time `json:"expires_at"`
	Issuer       string    `json:"issuer"`
	SerialNumber string    `json:"serial_number"`
	Fingerprint  string    `json:"fingerprint"`
	ACMEServer   string    `json:"acme_server"`
	Version      string    `json:"version"`
	RenewalCount int       `json:"renewal_count"`
}

// SaveCertificateMetadata saves certificate metadata to JSON file
func SaveCertificateMetadata(filePath string, metadata CertificateMetadata) error {
	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write metadata file: %w", err)
	}

	return nil
}

// LoadCertificateMetadata loads certificate metadata from JSON file
func LoadCertificateMetadata(filePath string) (CertificateMetadata, error) {
	var metadata CertificateMetadata

	data, err := os.ReadFile(filePath)
	if err != nil {
		return metadata, fmt.Errorf("failed to read metadata file: %w", err)
	}

	if err := json.Unmarshal(data, &metadata); err != nil {
		return metadata, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return metadata, nil
}

// UpdateRenewalCount increments the renewal count in metadata
func UpdateRenewalCount(filePath string) error {
	metadata, err := LoadCertificateMetadata(filePath)
	if err != nil {
		// If metadata doesn't exist, create new with count 1
		return nil
	}

	metadata.RenewalCount++
	return SaveCertificateMetadata(filePath, metadata)
}
