package dns

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/cloudflare/cloudflare-go"
	"github.com/go-acme/lego/v4/challenge/dns01"
)

// CloudflareProvider implements the DNS provider for Cloudflare
type CloudflareProvider struct {
	client  *cloudflare.API
	timeout time.Duration
	verbose bool
	records map[string]string // Track created records for cleanup
}

// NewCloudflareProvider creates a new Cloudflare DNS provider
func NewCloudflareProvider(apiToken, email string, timeout int, verbose bool) (*CloudflareProvider, error) {
	// Create Cloudflare API client
	api, err := cloudflare.NewWithAPIToken(apiToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create Cloudflare client: %w", err)
	}

	// Verify API token works
	ctx := context.Background()
	_, err = api.VerifyAPIToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("invalid Cloudflare API token: %w", err)
	}

	if verbose {
		log.Println("Cloudflare API client initialized successfully")
	}

	return &CloudflareProvider{
		client:  api,
		timeout: time.Duration(timeout) * time.Second,
		verbose: verbose,
		records: make(map[string]string),
	}, nil
}

// Present creates the DNS TXT record for the ACME challenge
func (p *CloudflareProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	if p.verbose {
		log.Printf("Creating DNS TXT record: %s = %s", fqdn, value)
	}

	// Extract the zone name from the domain and get zone ID
	zoneID, err := p.GetZoneIDForDomain(domain)
	if err != nil {
		return fmt.Errorf("failed to determine zone for domain %s: %w", domain, err)
	}

	ctx := context.Background()

	// Create the DNS record
	recordName := strings.TrimSuffix(fqdn, ".")
	createParams := cloudflare.CreateDNSRecordParams{
		Type:    "TXT",
		Name:    recordName,
		Content: value,
		TTL:     60, // Short TTL for quick propagation
	}

	response, err := p.client.CreateDNSRecord(ctx, cloudflare.ZoneIdentifier(zoneID), createParams)
	if err != nil {
		return fmt.Errorf("failed to create DNS record: %w", err)
	}

	// Store record ID for cleanup
	p.records[token] = response.ID

	if p.verbose {
		log.Printf("DNS record created successfully: %s", response.ID)
		log.Printf("Waiting for DNS propagation (up to %v)...", p.timeout)
	}

	// Simple sleep for DNS propagation - more reliable than complex checking
	time.Sleep(30 * time.Second)

	return nil
} // CleanUp removes the DNS TXT record after the challenge is complete
func (p *CloudflareProvider) CleanUp(domain, token, keyAuth string) error {
	recordID, exists := p.records[token]
	if !exists {
		if p.verbose {
			log.Printf("No record ID found for token %s, skipping cleanup", token)
		}
		return nil
	}

	if p.verbose {
		log.Printf("Cleaning up DNS record: %s", recordID)
	}

	// Extract the zone name from the domain and get zone ID
	zoneID, err := p.GetZoneIDForDomain(domain)
	if err != nil {
		return fmt.Errorf("failed to determine zone for domain %s: %w", domain, err)
	}

	ctx := context.Background()

	// Delete the DNS record
	err = p.client.DeleteDNSRecord(ctx, cloudflare.ZoneIdentifier(zoneID), recordID)
	if err != nil {
		return fmt.Errorf("failed to delete DNS record: %w", err)
	}

	// Remove from tracking
	delete(p.records, token)

	if p.verbose {
		log.Printf("DNS record cleaned up successfully")
	}

	return nil
}

// Timeout returns the timeout duration for DNS propagation
func (p *CloudflareProvider) Timeout() (timeout, interval time.Duration) {
	return p.timeout, 10 * time.Second
}
