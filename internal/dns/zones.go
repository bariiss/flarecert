package dns

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// ZoneInfo holds zone information
type ZoneInfo struct {
	ID     string
	Name   string
	Status string
}

// ListZones lists all available zones for the user to choose from
func (p *CloudflareProvider) ListZones() ([]ZoneInfo, error) {
	ctx := context.Background()
	zones, err := p.client.ListZones(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list zones: %w", err)
	}

	var zoneInfos []ZoneInfo
	for _, zone := range zones {
		zoneInfos = append(zoneInfos, ZoneInfo{
			ID:     zone.ID,
			Name:   zone.Name,
			Status: zone.Status,
		})
	}

	return zoneInfos, nil
}

// SelectZoneInteractive allows user to select a zone interactively
func (p *CloudflareProvider) SelectZoneInteractive(domain string) (string, error) {
	zones, err := p.ListZones()
	if err != nil {
		return "", err
	}

	if len(zones) == 0 {
		return "", fmt.Errorf("no zones found in your Cloudflare account")
	}

	// Try to find matching zones for the domain
	var matchingZones []ZoneInfo
	for _, zone := range zones {
		if strings.HasSuffix(domain, zone.Name) || domain == zone.Name {
			matchingZones = append(matchingZones, zone)
		}
	}

	// If we have exact matches, use them
	if len(matchingZones) > 0 {
		zones = matchingZones
	}

	if len(zones) == 1 {
		if p.verbose {
			fmt.Printf("🎯 Automatically selected zone: %s (%s)\n", zones[0].Name, zones[0].ID)
		}
		return zones[0].ID, nil
	}

	// Multiple zones, let user choose
	fmt.Printf("\n📋 Available Cloudflare zones for domain '%s':\n\n", domain)
	for i, zone := range zones {
		status := "✅"
		if zone.Status != "active" {
			status = "⚠️"
		}
		fmt.Printf("  %d. %s %s (%s)\n", i+1, status, zone.Name, zone.Status)
	}

	fmt.Print("\n🔢 Please select a zone (enter number): ")

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read input: %w", err)
	}

	input = strings.TrimSpace(input)
	choice, err := strconv.Atoi(input)
	if err != nil {
		return "", fmt.Errorf("invalid choice: %s", input)
	}

	if choice < 1 || choice > len(zones) {
		return "", fmt.Errorf("choice out of range: %d", choice)
	}

	selectedZone := zones[choice-1]
	fmt.Printf("✅ Selected zone: %s (%s)\n\n", selectedZone.Name, selectedZone.ID)

	return selectedZone.ID, nil
}

// GetZoneIDForDomain gets the zone ID for a domain, with interactive selection if needed
func (p *CloudflareProvider) GetZoneIDForDomain(domain string) (string, error) {
	// First try automatic detection
	zoneID, err := p.getZoneIDAutomatic(domain)
	if err == nil {
		return zoneID, nil
	}

	if p.verbose {
		fmt.Printf("⚠️  Automatic zone detection failed: %v\n", err)
		fmt.Printf("🔍 Switching to interactive zone selection...\n")
	}

	// Fall back to interactive selection
	return p.SelectZoneInteractive(domain)
}

// getZoneIDAutomatic tries to automatically detect the zone for a domain
func (p *CloudflareProvider) getZoneIDAutomatic(domain string) (string, error) {
	// Remove any subdomain parts to find the zone
	parts := strings.Split(domain, ".")
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid domain: %s", domain)
	}

	// Try different combinations to find the zone
	ctx := context.Background()
	for i := 0; i < len(parts)-1; i++ {
		candidateZone := strings.Join(parts[i:], ".")

		zones, err := p.client.ListZones(ctx, candidateZone)
		if err != nil {
			continue
		}

		if len(zones) > 0 {
			if p.verbose {
				fmt.Printf("🎯 Found zone automatically: %s (%s)\n", zones[0].Name, zones[0].ID)
			}
			return zones[0].ID, nil
		}
	}

	return "", fmt.Errorf("no zone found for domain: %s", domain)
}
