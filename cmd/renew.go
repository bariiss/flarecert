package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/bariiss/flarecert/internal/acme"
	"github.com/bariiss/flarecert/internal/certificate"

	"github.com/spf13/cobra"
)

var renewCmd = &cobra.Command{
	Use:   "renew",
	Short: "Renew existing SSL certificates",
	Long: `Renew SSL certificates that are close to expiration.

This command will scan the certificate directory and renew any certificates
that expire within the next 30 days.`,
	RunE: runRenewCommand,
}

var (
	renewDays    int
	renewCertDir string
	renewAll     bool
)

func init() {
	rootCmd.AddCommand(renewCmd)

	renewCmd.Flags().IntVar(&renewDays, "days", 30, "Renew certificates expiring within this many days")
	renewCmd.Flags().StringVar(&renewCertDir, "cert-dir", "./certs", "Directory containing certificates")
	renewCmd.Flags().BoolVar(&renewAll, "all", false, "Renew all certificates regardless of expiration")
}

func runRenewCommand(cmd *cobra.Command, args []string) error {
	verbose, _ := cmd.Flags().GetBool("verbose")

	if verbose {
		log.Println("Starting certificate renewal check...")
	}

	// Find certificates to renew
	certsToRenew, err := findCertificatesForRenewal(renewCertDir, renewDays, renewAll, verbose)
	if err != nil {
		return fmt.Errorf("failed to find certificates for renewal: %w", err)
	}

	if len(certsToRenew) == 0 {
		fmt.Println("✅ No certificates need renewal")
		return nil
	}

	fmt.Printf("Found %d certificate(s) to renew:\n", len(certsToRenew))
	for _, cert := range certsToRenew {
		fmt.Printf("  - %s (expires: %s)\n", cert.Domain, cert.ExpiresAt.Format("2006-01-02"))
	}

	// Renew each certificate
	for _, cert := range certsToRenew {
		fmt.Printf("\n🔄 Renewing certificate for: %s\n", cert.Domain)

		// Create certificate manager for renewal (force renew enabled)
		manager, err := certificate.NewManager(renewCertDir, "rsa2048", false, true, verbose)
		if err != nil {
			log.Printf("❌ Failed to create certificate manager for %s: %v", cert.Domain, err)
			continue
		}

		// Renew certificate
		if err := manager.GenerateCertificate(cert.Domains); err != nil {
			log.Printf("❌ Failed to renew %s: %v", cert.Domain, err)
			continue
		}

		fmt.Printf("✅ Successfully renewed certificate for %s\n", cert.Domain)
	}

	return nil
}

type CertificateInfo struct {
	Domain    string
	Domains   []string
	ExpiresAt time.Time
	Path      string
}

func findCertificatesForRenewal(certDir string, days int, renewAll bool, verbose bool) ([]CertificateInfo, error) {
	var certificates []CertificateInfo
	threshold := time.Now().AddDate(0, 0, days)

	entries, err := os.ReadDir(certDir)
	if err != nil {
		if os.IsNotExist(err) {
			return certificates, nil
		}
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		domainName := entry.Name()
		// Use the new certificate structure with current subdirectory
		certPath := filepath.Join(certDir, domainName, "current", "cert.pem")

		if _, err := os.Stat(certPath); os.IsNotExist(err) {
			if verbose {
				log.Printf("Skipping %s: no cert.pem found in current directory", domainName)
			}
			continue
		}

		// Parse certificate to get expiration and domains
		domains, expiresAt, err := acme.ParseCertificateInfo(certPath)
		if err != nil {
			if verbose {
				log.Printf("Skipping %s: failed to parse certificate: %v", domainName, err)
			}
			continue
		}

		// Check if renewal is needed
		if renewAll || expiresAt.Before(threshold) {
			certificates = append(certificates, CertificateInfo{
				Domain:    domainName,
				Domains:   domains,
				ExpiresAt: expiresAt,
				Path:      certPath,
			})
		} else if verbose {
			log.Printf("Certificate %s is valid until %s (no renewal needed)", domainName, expiresAt.Format("2006-01-02"))
		}
	}

	return certificates, nil
}
