package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/bariiss/flarecert/internal/acme"
	"github.com/bariiss/flarecert/internal/k8s"
	"github.com/bariiss/flarecert/internal/utils"

	"github.com/spf13/cobra"
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export existing certificates to Kubernetes Secret YAML",
	Long: `Export existing SSL certificates to Kubernetes Secret YAML format.

This command scans existing certificates and generates Kubernetes Secret YAML files
for them. Useful when you want to create K8s secrets for certificates that were
generated without the --k8s flag.

Examples:
  # Export all certificates to K8s secrets
  flarecert export --all

  # Export specific certificate
  flarecert export --domain example.com

  # Export with custom output directory
  flarecert export --domain example.com --output ./k8s-secrets/`,
	RunE: runExportCommand,
}

var (
	exportDomain    string
	exportAll       bool
	exportCertDir   string
	exportOutputDir string
)

func init() {
	rootCmd.AddCommand(exportCmd)

	exportCmd.Flags().StringVar(&exportDomain, "domain", "", "Domain name to export (if not specified with --all, will list available certificates)")
	exportCmd.Flags().BoolVar(&exportAll, "all", false, "Export all available certificates")
	exportCmd.Flags().StringVar(&exportCertDir, "cert-dir", "./certs", "Directory containing certificates")
	exportCmd.Flags().StringVar(&exportOutputDir, "output", "", "Output directory for YAML files (default: same as certificate directory)")

	// Register completion for domain flag
	exportCmd.RegisterFlagCompletionFunc("domain", GetDomainCompletions)
}

func runExportCommand(cmd *cobra.Command, args []string) error {
	verbose, _ := cmd.Flags().GetBool("verbose")

	if verbose {
		log.Println("Starting certificate export...")
	}

	// Validate flags
	if !exportAll && exportDomain == "" {
		// No domain specified and not --all, show available certificates
		return showAvailableCertificates()
	}

	if exportAll && exportDomain != "" {
		return fmt.Errorf("cannot use both --all and --domain flags together")
	}

	// Find certificates to export
	var certsToExport []CertificateExportInfo
	var err error

	if exportAll {
		certsToExport, err = findAllCertificates(exportCertDir, verbose)
		if err != nil {
			return fmt.Errorf("failed to find certificates: %w", err)
		}
	} else {
		cert, err := findCertificateByDomain(exportCertDir, exportDomain, verbose)
		if err != nil {
			return fmt.Errorf("failed to find certificate for domain %s: %w", exportDomain, err)
		}
		certsToExport = []CertificateExportInfo{*cert}
	}

	if len(certsToExport) == 0 {
		fmt.Println("No certificates found to export")
		return nil
	}

	// Export certificates
	secretGen := k8s.NewSecretGenerator(verbose)
	successCount := 0

	for _, cert := range certsToExport {
		fmt.Printf("📝 Exporting certificate: %s\n", cert.DirectoryName)

		// Determine output directory
		outputDir := exportOutputDir
		if outputDir == "" {
			outputDir = cert.CertificateDir
		}

		// Create custom paths for output
		outputPaths := utils.CertificatePaths{
			CertDir:       outputDir,
			CurrentDir:    outputDir,
			ArchiveDir:    filepath.Join(cert.CertificateDir, "archive"),
			LogsDir:       filepath.Join(cert.CertificateDir, "logs"),
			CertFile:      cert.CertFile,
			KeyFile:       cert.KeyFile,
			ChainFile:     cert.ChainFile,
			FullchainFile: cert.FullchainFile,
			InfoFile:      cert.InfoFile,
		}

		// Determine primary domain (prefer wildcard)
		primaryDomain := cert.Domains[0]
		for _, domain := range cert.Domains {
			if strings.HasPrefix(domain, "*.") {
				primaryDomain = domain
				break
			}
		}

		if err := secretGen.CreateSecret(&outputPaths, primaryDomain, cert.Domains); err != nil {
			log.Printf("❌ Failed to export %s: %v", cert.DirectoryName, err)
			continue
		}

		successCount++
	}

	fmt.Printf("\n✅ Successfully exported %d/%d certificate(s) to Kubernetes Secret YAML\n",
		successCount, len(certsToExport))

	return nil
}

type CertificateExportInfo struct {
	DirectoryName  string
	CertificateDir string
	Domains        []string
	ExpiresAt      string
	CertFile       string
	KeyFile        string
	ChainFile      string
	FullchainFile  string
	InfoFile       string
}

func findAllCertificates(certDir string, verbose bool) ([]CertificateExportInfo, error) {
	var certificates []CertificateExportInfo

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

		cert, err := buildCertificateExportInfo(certDir, entry.Name(), verbose)
		if err != nil {
			if verbose {
				log.Printf("Skipping %s: %v", entry.Name(), err)
			}
			continue
		}

		certificates = append(certificates, *cert)
	}

	return certificates, nil
}

func findCertificateByDomain(certDir, domain string, verbose bool) (*CertificateExportInfo, error) {
	// Try to find certificate directory by domain
	entries, err := os.ReadDir(certDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		cert, err := buildCertificateExportInfo(certDir, entry.Name(), verbose)
		if err != nil {
			continue
		}

		// Check if any of the certificate domains match
		for _, certDomain := range cert.Domains {
			if certDomain == domain {
				return cert, nil
			}
		}
	}

	return nil, fmt.Errorf("certificate not found for domain: %s", domain)
}

func buildCertificateExportInfo(certDir, dirName string, verbose bool) (*CertificateExportInfo, error) {
	certPath := filepath.Join(certDir, dirName, "current", "cert.pem")
	keyPath := filepath.Join(certDir, dirName, "current", "privkey.pem")
	chainPath := filepath.Join(certDir, dirName, "current", "chain.pem")
	fullchainPath := filepath.Join(certDir, dirName, "current", "fullchain.pem")
	infoPath := filepath.Join(certDir, dirName, "current", "cert.json")

	// Check if certificate exists
	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("certificate file not found: %s", certPath)
	}

	// Check if key exists
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("private key file not found: %s", keyPath)
	}

	// Parse certificate to get domains and expiration
	domains, expiresAt, err := acme.ParseCertificateInfo(certPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	return &CertificateExportInfo{
		DirectoryName:  dirName,
		CertificateDir: filepath.Join(certDir, dirName),
		Domains:        domains,
		ExpiresAt:      expiresAt.Format("2006-01-02 15:04"),
		CertFile:       certPath,
		KeyFile:        keyPath,
		ChainFile:      chainPath,
		FullchainFile:  fullchainPath,
		InfoFile:       infoPath,
	}, nil
}

func showAvailableCertificates() error {
	fmt.Println("Available certificates to export:")
	fmt.Println()

	certificates, err := findAllCertificates(exportCertDir, false)
	if err != nil {
		return fmt.Errorf("failed to find certificates: %w", err)
	}

	if len(certificates) == 0 {
		fmt.Println("No certificates found in the certificate directory.")
		fmt.Printf("Certificate directory: %s\n", exportCertDir)
		return nil
	}

	fmt.Printf("%-25s %-35s %s\n", "DIRECTORY", "DOMAINS", "EXPIRES")
	fmt.Printf("%-25s %-35s %s\n", "---------", "-------", "-------")

	for _, cert := range certificates {
		domainsStr := strings.Join(cert.Domains, ", ")
		if len(domainsStr) > 33 {
			domainsStr = domainsStr[:30] + "..."
		}

		fmt.Printf("%-25s %-35s %s\n", cert.DirectoryName, domainsStr, cert.ExpiresAt)
	}

	fmt.Println()
	fmt.Println("Usage:")
	fmt.Printf("  flarecert export --domain <domain>  # Export specific certificate\n")
	fmt.Printf("  flarecert export --all              # Export all certificates\n")

	return nil
}
