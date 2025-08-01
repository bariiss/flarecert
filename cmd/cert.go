package cmd

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/bariiss/flarecert/internal/acme"
	"github.com/bariiss/flarecert/internal/config"
	"github.com/bariiss/flarecert/internal/utils"

	"github.com/spf13/cobra"
)

var certCmd = &cobra.Command{
	Use:   "cert",
	Short: "Generate a new SSL certificate",
	Long: `Generate a new SSL certificate using Let's Encrypt with Cloudflare DNS-01 challenge.

Examples:
  # Single domain
  flarecert cert --domain example.com

  # Wildcard domain
  flarecert cert --domain "*.example.com"

  # Multiple domains (SAN certificate)
  flarecert cert --domain example.com --domain www.example.com --domain api.example.com`,
	RunE: runCertCommand,
}

var (
	domains    []string
	certDir    string
	staging    bool
	keyType    string
	forceRenew bool
)

func init() {
	rootCmd.AddCommand(certCmd)

	certCmd.Flags().StringSliceVarP(&domains, "domain", "d", []string{}, "Domain name(s) for the certificate (required)")
	certCmd.Flags().StringVar(&certDir, "cert-dir", "./certs", "Directory to store certificates")
	certCmd.Flags().BoolVar(&staging, "staging", false, "Use Let's Encrypt staging environment")
	certCmd.Flags().StringVar(&keyType, "key-type", "rsa2048", "Key type: rsa2048, rsa4096, ec256, ec384")
	certCmd.Flags().BoolVar(&forceRenew, "force", false, "Force renewal even if certificate is valid")

	// Register completion for domain flag
	certCmd.RegisterFlagCompletionFunc("domain", GetDomainCompletions)

	// Register completion for key-type flag
	certCmd.RegisterFlagCompletionFunc("key-type", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"rsa2048", "rsa4096", "ec256", "ec384"}, cobra.ShellCompDirectiveNoFileComp
	})

	certCmd.MarkFlagRequired("domain")
}

func runCertCommand(cmd *cobra.Command, args []string) error {
	verbose, _ := cmd.Flags().GetBool("verbose")

	if verbose {
		log.Println("Starting certificate generation...")
	}

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

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Override staging if flag is set
	if staging {
		cfg.ACMEServer = "https://acme-staging-v02.api.letsencrypt.org/directory"
	}

	// Create certificate directory structure
	primaryDomain := domains[0]
	if err := utils.CreateCertificateStructure(certDir, primaryDomain); err != nil {
		return fmt.Errorf("failed to create certificate structure: %w", err)
	}

	// Get certificate paths
	paths := utils.GetCertificatePaths(certDir, primaryDomain)

	// Initialize ACME client
	client, err := acme.NewClient(cfg, verbose)
	if err != nil {
		return fmt.Errorf("failed to create ACME client: %w", err)
	}

	// Check if certificate already exists and is valid
	primaryDomain = domains[0]
	certPath := paths.CertFile

	if !forceRenew {
		if valid, err := acme.IsCertificateValid(certPath, domains); err == nil && valid {
			fmt.Printf("Certificate for %s is already valid and not expired\n", strings.Join(domains, ", "))
			fmt.Printf("Use --force to renew anyway\n")
			return nil
		}
	}

	// Archive old certificate if it exists
	if err := utils.ArchiveOldCertificate(paths); err != nil {
		if verbose {
			log.Printf("Warning: failed to archive old certificate: %v", err)
		}
	}

	// Generate certificate
	fmt.Printf("🔐 Generating certificate for: %s\n", utils.FormatDomainForDisplay(domains))

	cert, err := client.ObtainCertificate(domains, keyType)
	if err != nil {
		return fmt.Errorf("failed to obtain certificate: %w", err)
	}

	// Save certificate files with improved structure
	files := map[string][]byte{
		paths.CertFile:      cert.Certificate,
		paths.KeyFile:       cert.PrivateKey,
		paths.ChainFile:     cert.IssuerCertificate,
		paths.FullchainFile: append(cert.Certificate, cert.IssuerCertificate...),
	}

	for filename, data := range files {
		if err := os.WriteFile(filename, data, 0600); err != nil {
			return fmt.Errorf("failed to save %s: %w", filename, err)
		}
		if verbose {
			log.Printf("💾 Saved: %s", filename)
		}
	}

	// Save certificate metadata
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
		KeyType:      keyType,
		CreatedAt:    time.Now(),
		ExpiresAt:    cert.NotAfter,
		ACMEServer:   cfg.ACMEServer,
		Version:      "1.0",
		RenewalCount: 0,
	}

	if err := utils.SaveCertificateMetadata(paths.InfoFile, metadata); err != nil {
		if verbose {
			log.Printf("Warning: failed to save metadata: %v", err)
		}
	}

	// Cleanup old archives (keep last 30 days)
	if err := utils.CleanupOldArchives(paths.ArchiveDir, 30); err != nil {
		if verbose {
			log.Printf("Warning: failed to cleanup old archives: %v", err)
		}
	}

	fmt.Printf("✅ Certificate successfully generated and saved to: %s\n", paths.CurrentDir)
	fmt.Printf("📅 Certificate expires: %s\n", cert.NotAfter.Format("2006-01-02 15:04:05 MST"))

	return nil
}
