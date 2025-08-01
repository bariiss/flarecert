package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// CertificateInfo holds information about a certificate
type CertificateInfo struct {
	Domain        string
	Domains       []string
	IsWildcard    bool
	ExpiresAt     time.Time
	CertPath      string
	KeyPath       string
	ChainPath     string
	FullchainPath string
	CreatedAt     time.Time
	KeyType       string
}

// GetCertificateDir returns the directory path for a certificate
func GetCertificateDir(baseDir, domain string) string {
	// Clean domain name for directory
	cleanDomain := strings.ReplaceAll(domain, "*.", "wildcard.")
	cleanDomain = strings.ReplaceAll(cleanDomain, "*", "wildcard")

	return filepath.Join(baseDir, cleanDomain)
}

// CreateCertificateStructure creates the directory structure for certificates
func CreateCertificateStructure(baseDir, domain string) error {
	certDir := GetCertificateDir(baseDir, domain)

	// Create main certificate directory
	if err := os.MkdirAll(certDir, 0755); err != nil {
		return fmt.Errorf("failed to create certificate directory: %w", err)
	}

	// Create subdirectories for better organization
	subdirs := []string{
		filepath.Join(certDir, "current"), // Current active certificate
		filepath.Join(certDir, "archive"), // Previous certificates
		filepath.Join(certDir, "logs"),    // Certificate generation logs
	}

	for _, dir := range subdirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

// GetCertificatePaths returns all the file paths for a certificate
func GetCertificatePaths(baseDir, domain string) CertificatePaths {
	certDir := GetCertificateDir(baseDir, domain)
	currentDir := filepath.Join(certDir, "current")

	return CertificatePaths{
		CertDir:       certDir,
		CurrentDir:    currentDir,
		ArchiveDir:    filepath.Join(certDir, "archive"),
		LogsDir:       filepath.Join(certDir, "logs"),
		CertFile:      filepath.Join(currentDir, "cert.pem"),
		KeyFile:       filepath.Join(currentDir, "privkey.pem"),
		ChainFile:     filepath.Join(currentDir, "chain.pem"),
		FullchainFile: filepath.Join(currentDir, "fullchain.pem"),
		InfoFile:      filepath.Join(currentDir, "cert.json"),
	}
}

// CertificatePaths holds all file paths for a certificate
type CertificatePaths struct {
	CertDir       string
	CurrentDir    string
	ArchiveDir    string
	LogsDir       string
	CertFile      string
	KeyFile       string
	ChainFile     string
	FullchainFile string
	InfoFile      string
}

// ArchiveOldCertificate moves current certificate to archive before creating new one
func ArchiveOldCertificate(paths CertificatePaths) error {
	// Check if current certificate exists
	if _, err := os.Stat(paths.CertFile); os.IsNotExist(err) {
		return nil // No current certificate to archive
	}

	// Create archive filename with timestamp
	timestamp := time.Now().Format("20060102-150405")
	archivePrefix := fmt.Sprintf("cert-%s", timestamp)

	// Files to archive
	files := map[string]string{
		paths.CertFile:      filepath.Join(paths.ArchiveDir, archivePrefix+"-cert.pem"),
		paths.KeyFile:       filepath.Join(paths.ArchiveDir, archivePrefix+"-privkey.pem"),
		paths.ChainFile:     filepath.Join(paths.ArchiveDir, archivePrefix+"-chain.pem"),
		paths.FullchainFile: filepath.Join(paths.ArchiveDir, archivePrefix+"-fullchain.pem"),
		paths.InfoFile:      filepath.Join(paths.ArchiveDir, archivePrefix+"-cert.json"),
	}

	for source, dest := range files {
		if _, err := os.Stat(source); err == nil {
			if err := os.Rename(source, dest); err != nil {
				return fmt.Errorf("failed to archive %s: %w", source, err)
			}
		}
	}

	return nil
}

// CleanupOldArchives removes archive files older than specified days
func CleanupOldArchives(archiveDir string, keepDays int) error {
	if keepDays <= 0 {
		return nil // Don't cleanup if keepDays is 0 or negative
	}

	cutoff := time.Now().AddDate(0, 0, -keepDays)

	entries, err := os.ReadDir(archiveDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Archive directory doesn't exist
		}
		return fmt.Errorf("failed to read archive directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		if info.ModTime().Before(cutoff) {
			filePath := filepath.Join(archiveDir, entry.Name())
			if err := os.Remove(filePath); err != nil {
				// Log error but continue cleanup
				fmt.Printf("Warning: failed to remove old archive file %s: %v\n", filePath, err)
			}
		}
	}

	return nil
}

// FormatDomainForDisplay formats domain names for user-friendly display
func FormatDomainForDisplay(domains []string) string {
	if len(domains) == 0 {
		return ""
	}

	if len(domains) == 1 {
		return domains[0]
	}

	// Show primary domain + count of additional domains
	additional := len(domains) - 1
	return fmt.Sprintf("%s (+%d more)", domains[0], additional)
}

// ValidateDomainName performs basic domain name validation
func ValidateDomainName(domain string) error {
	if domain == "" {
		return fmt.Errorf("domain cannot be empty")
	}

	// Allow wildcard domains
	if strings.HasPrefix(domain, "*.") {
		domain = domain[2:] // Remove *. prefix for validation
	}

	// Basic validation
	if strings.Contains(domain, " ") {
		return fmt.Errorf("domain cannot contain spaces: %s", domain)
	}

	parts := strings.Split(domain, ".")
	if len(parts) < 2 {
		return fmt.Errorf("domain must have at least two parts: %s", domain)
	}

	for _, part := range parts {
		if part == "" {
			return fmt.Errorf("domain parts cannot be empty: %s", domain)
		}
		if len(part) > 63 {
			return fmt.Errorf("domain part too long (max 63 chars): %s", part)
		}
	}

	return nil
}
