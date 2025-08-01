package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds the application configuration
type Config struct {
	CloudflareAPIToken string
	CloudflareEmail    string
	ACMEEmail          string
	ACMEServer         string
	CertDir            string
	DNSTimeout         int
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		CloudflareAPIToken: os.Getenv("CLOUDFLARE_API_TOKEN"),
		CloudflareEmail:    os.Getenv("CLOUDFLARE_EMAIL"),
		ACMEEmail:          os.Getenv("ACME_EMAIL"),
		ACMEServer:         os.Getenv("ACME_SERVER"),
		CertDir:            os.Getenv("CERT_DIR"),
		DNSTimeout:         300, // default 5 minutes
	}

	// Set defaults
	if cfg.ACMEServer == "" {
		cfg.ACMEServer = "https://acme-v02.api.letsencrypt.org/directory"
	}

	if cfg.CertDir == "" {
		cfg.CertDir = "./certs"
	}

	// Parse DNS timeout
	if timeoutStr := os.Getenv("DNS_PROPAGATION_TIMEOUT"); timeoutStr != "" {
		if timeout, err := strconv.Atoi(timeoutStr); err == nil && timeout > 0 {
			cfg.DNSTimeout = timeout
		}
	}

	// Validate required fields
	if cfg.CloudflareAPIToken == "" {
		return nil, fmt.Errorf("CLOUDFLARE_API_TOKEN is required")
	}

	if cfg.CloudflareEmail == "" {
		return nil, fmt.Errorf("CLOUDFLARE_EMAIL is required")
	}

	if cfg.ACMEEmail == "" {
		return nil, fmt.Errorf("ACME_EMAIL is required")
	}

	return cfg, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.CloudflareAPIToken == "" {
		return fmt.Errorf("cloudflare API token is required")
	}

	if c.CloudflareEmail == "" {
		return fmt.Errorf("cloudflare email is required")
	}

	if c.ACMEEmail == "" {
		return fmt.Errorf("ACME email is required")
	}

	if c.ACMEServer == "" {
		return fmt.Errorf("ACME server URL is required")
	}

	return nil
}
