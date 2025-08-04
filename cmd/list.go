package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/bariiss/flarecert/internal/acme"
	"github.com/bariiss/flarecert/internal/config"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List existing SSL certificates",
	Long: `List all SSL certificates in the certificate directory with their expiration dates.

This command scans the certificate directory and displays information about
all stored certificates including their domains and expiration dates.`,
	RunE: runListCommand,
}

var listCertDir string

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().StringVar(&listCertDir, "cert-dir", "./certs", "Directory containing certificates")
}

func runListCommand(cmd *cobra.Command, args []string) error {
	verbose, _ := cmd.Flags().GetBool("verbose")

	// Load configuration to get the correct cert directory
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Use cert directory from config if not overridden by flag
	certDir := listCertDir
	if certDir == "./certs" { // Default value means flag wasn't set
		certDir = cfg.CertDir
	}

	if verbose {
		log.Printf("Scanning certificate directory: %s", certDir)
	}

	// Expand tilde in path if present
	if strings.HasPrefix(certDir, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get user home directory: %w", err)
		}
		certDir = filepath.Join(homeDir, certDir[2:])
	}

	entries, err := os.ReadDir(certDir)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("No certificates found in the specified directory.\n")
			printEmptyTable()
			return nil
		}
		return fmt.Errorf("failed to read certificate directory: %w", err)
	}

	// Create table writer
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	fmt.Fprintln(w, "DOMAIN\tDOMAINS\tEXPIRATION\tSTATUS")
	fmt.Fprintln(w, "------\t-------\t----------\t------")

	found := false
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		domainName := entry.Name()
		certPath := filepath.Join(certDir, domainName, "current", "cert.pem")

		if _, err := os.Stat(certPath); os.IsNotExist(err) {
			if verbose {
				log.Printf("Skipping %s: no cert.pem found in current/", domainName)
			}
			continue
		}

		domains, expiresAt, err := acme.ParseCertificateInfo(certPath)
		if err != nil {
			if verbose {
				log.Printf("Skipping %s: failed to parse certificate: %v", domainName, err)
			}
			continue
		}

		found = true

		// Determine status
		status := "✅ Valid"
		if expiresAt.Before(time.Now()) {
			status = "❌ Expired"
		} else if expiresAt.Before(time.Now().AddDate(0, 0, 30)) {
			status = "⚠️  Expires Soon"
		}

		// Format domains
		domainsStr := strings.Join(domains, ", ")
		if len(domainsStr) > 40 {
			domainsStr = domainsStr[:37] + "..."
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			domainName,
			domainsStr,
			expiresAt.Format("2006-01-02 15:04"),
			status,
		)
	}

	if !found {
		fmt.Println("No certificates found in the specified directory.")
	}

	return nil
}

// printEmptyTable prints an empty table with headers
func printEmptyTable() {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	fmt.Fprintln(w, "DOMAIN\tDOMAINS\tEXPIRATION\tSTATUS")
	fmt.Fprintln(w, "------\t-------\t----------\t------")
}
