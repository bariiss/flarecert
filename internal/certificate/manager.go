package certificate

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/bariiss/flarecert/internal/acme"
	"github.com/bariiss/flarecert/internal/config"
	"github.com/bariiss/flarecert/internal/ui"
	"github.com/bariiss/flarecert/internal/utils"
)

// Manager handles certificate operations
type Manager struct {
	config     *config.Config
	certDir    string
	verbose    bool
	staging    bool
	keyType    string
	forceRenew bool
}

// NewManager creates a new certificate manager
func NewManager(certDir, keyType string, staging, forceRenew, verbose bool) (*Manager, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Override staging if flag is set
	if staging {
		cfg.ACMEServer = "https://acme-staging-v02.api.letsencrypt.org/directory"
	}

	return &Manager{
		config:     cfg,
		certDir:    certDir,
		verbose:    verbose,
		staging:    staging,
		keyType:    keyType,
		forceRenew: forceRenew,
	}, nil
}

// GenerateCertificate generates a new certificate for the given domains
func (m *Manager) GenerateCertificate(domains []string) error {
	// Validate input
	if len(domains) == 0 {
		return fmt.Errorf("at least one domain must be specified")
	}

	// Validate each domain
	for _, domain := range domains {
		if err := utils.ValidateDomainName(domain); err != nil {
			return fmt.Errorf("invalid domain %s: %w", domain, err)
		}
	}

	// Create certificate directory structure using domain list (prioritizes wildcard)
	if err := utils.CreateCertificateStructureForDomains(m.certDir, domains); err != nil {
		return fmt.Errorf("failed to create certificate structure: %w", err)
	}

	// Get certificate paths using domain list (prioritizes wildcard)
	paths := utils.GetCertificatePathsForDomains(m.certDir, domains)

	// Initialize ACME client
	client, err := acme.NewClient(m.config, m.verbose, m.keyType)
	if err != nil {
		return fmt.Errorf("failed to create ACME client: %w", err)
	}

	// Check existing certificate and determine action
	action, err := m.determineAction(domains, paths)
	if err != nil {
		return err
	}

	switch action {
	case ActionSkip:
		fmt.Printf("Certificate generation cancelled. Use --force to renew without prompting.\n")
		return nil
	case ActionRenew, ActionReplace:
		// Continue with certificate generation
	}

	// Archive old certificate if it exists
	if err := utils.ArchiveOldCertificate(paths); err != nil {
		if m.verbose {
			fmt.Printf("Warning: failed to archive old certificate: %v\n", err)
		}
	}

	// Generate certificate
	fmt.Printf("🔐 Generating certificate for: %s\n", utils.FormatDomainForDisplay(domains))

	cert, err := client.ObtainCertificate(domains)
	if err != nil {
		return fmt.Errorf("failed to obtain certificate: %w", err)
	}

	// Save certificate files
	if err := m.saveCertificateFiles(cert, paths); err != nil {
		return err
	}

	// Save certificate metadata
	if err := m.saveCertificateMetadata(domains, paths, cert); err != nil {
		if m.verbose {
			fmt.Printf("Warning: failed to save metadata: %v\n", err)
		}
	}

	// Cleanup old archives
	if err := utils.CleanupOldArchives(paths.ArchiveDir, 30); err != nil {
		if m.verbose {
			fmt.Printf("Warning: failed to cleanup old archives: %v\n", err)
		}
	}

	fmt.Printf("✅ Certificate successfully generated and saved to: %s\n", paths.CurrentDir)
	fmt.Printf("📅 Certificate expires: %s\n", cert.NotAfter.Format("2006-01-02 15:04:05 MST"))

	return nil
}

// CertificateAction represents the action to take for a certificate
type CertificateAction int

const (
	ActionSkip CertificateAction = iota
	ActionRenew
	ActionReplace
)

// determineAction determines what action to take based on existing certificate
func (m *Manager) determineAction(domains []string, paths utils.CertificatePaths) (CertificateAction, error) {
	if m.forceRenew {
		return ActionRenew, nil
	}

	// Check if certificate exists
	if _, err := os.Stat(paths.CertFile); os.IsNotExist(err) {
		return ActionRenew, nil // No existing certificate
	}

	// Parse existing certificate
	existingDomains, expiresAt, err := acme.ParseCertificateInfo(paths.CertFile)
	if err != nil {
		if m.verbose {
			fmt.Printf("Warning: failed to parse existing certificate: %v\n", err)
		}
		return ActionRenew, nil
	}

	// Check if domains match
	if DomainsMatch(domains, existingDomains) {
		now := time.Now()
		daysRemaining := int(expiresAt.Sub(now).Hours() / 24)

		if expiresAt.Before(now) {
			// Certificate is expired
			fmt.Printf("⚠️  Certificate for %s has expired (%d days ago)\n",
				strings.Join(domains, ", "), -daysRemaining)
			fmt.Printf("🔄 Automatically renewing expired certificate...\n")
			return ActionRenew, nil
		} else if daysRemaining <= 30 {
			// Certificate expires soon
			fmt.Printf("⚠️  Certificate for %s expires in %d days (%s)\n",
				strings.Join(domains, ", "), daysRemaining, expiresAt.Format("2006-01-02 15:04"))

			if !ui.AskUserConfirmation("Do you want to renew it now?") {
				return ActionSkip, nil
			}
			fmt.Printf("🔄 Renewing certificate...\n")
			return ActionRenew, nil
		} else {
			// Certificate is valid
			fmt.Printf("✅ Certificate for %s is already valid and expires in %d days (%s)\n",
				strings.Join(domains, ", "), daysRemaining, expiresAt.Format("2006-01-02 15:04"))

			if !ui.AskUserConfirmation("Do you want to renew it anyway?") {
				return ActionSkip, nil
			}
			fmt.Printf("🔄 Force renewing certificate...\n")
			return ActionRenew, nil
		}
	} else {
		// Different domains
		fmt.Printf("⚠️  Found existing certificate with different domains:\n")
		fmt.Printf("   Existing: %s\n", strings.Join(existingDomains, ", "))
		fmt.Printf("   Requested: %s\n", strings.Join(domains, ", "))

		if !ui.AskUserConfirmation("Do you want to replace it with the new certificate?") {
			return ActionSkip, nil
		}
		fmt.Printf("🔄 Replacing certificate with new domains...\n")
		return ActionReplace, nil
	}
}

// saveCertificateFiles saves all certificate files to disk
func (m *Manager) saveCertificateFiles(cert *acme.CertificateResult, paths utils.CertificatePaths) error {
	// Trim leading/trailing whitespace from IssuerCertificate and ensure proper formatting
	trimmedIssuerCert := strings.TrimSpace(string(cert.IssuerCertificate))
	if trimmedIssuerCert != "" && !strings.HasSuffix(trimmedIssuerCert, "\n") {
		trimmedIssuerCert += "\n"
	}

	files := map[string][]byte{
		paths.CertFile:      cert.Certificate,
		paths.KeyFile:       cert.PrivateKey,
		paths.ChainFile:     []byte(trimmedIssuerCert),
		paths.FullchainFile: append(cert.Certificate, cert.IssuerCertificate...),
	}

	for filename, data := range files {
		if err := os.WriteFile(filename, data, 0600); err != nil {
			return fmt.Errorf("failed to save %s: %w", filename, err)
		}
		if m.verbose {
			fmt.Printf("💾 Saved: %s\n", filename)
		}
	}

	return nil
}

// saveCertificateMetadata saves certificate metadata to disk
func (m *Manager) saveCertificateMetadata(domains []string, paths utils.CertificatePaths, cert *acme.CertificateResult) error {
	primaryDomain := domains[0]
	isWildcard := false
	for _, domain := range domains {
		if strings.HasPrefix(domain, "*.") {
			isWildcard = true
			break
		}
	}

	metadata := utils.CertificateMetadata{
		Domain:       primaryDomain,
		Domains:      domains,
		IsWildcard:   isWildcard,
		KeyType:      m.keyType,
		CreatedAt:    time.Now(),
		ExpiresAt:    cert.NotAfter,
		ACMEServer:   m.config.ACMEServer,
		Version:      "1.0",
		RenewalCount: 0,
	}

	return utils.SaveCertificateMetadata(paths.InfoFile, metadata)
}

// DomainsMatch checks if two domain slices contain the same domains (ignoring duplicates)
func DomainsMatch(domains1, domains2 []string) bool {
	// Create sets to check for equality regardless of order and duplicates
	set1 := make(map[string]bool)
	set2 := make(map[string]bool)

	for _, domain := range domains1 {
		set1[strings.ToLower(strings.TrimSpace(domain))] = true
	}

	for _, domain := range domains2 {
		set2[strings.ToLower(strings.TrimSpace(domain))] = true
	}

	// Check if sets have the same size
	if len(set1) != len(set2) {
		return false
	}

	// Check if all domains in set1 exist in set2
	for domain := range set1 {
		if !set2[domain] {
			return false
		}
	}

	return true
}
