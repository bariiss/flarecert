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

	if verbose {
		log.Println("Scanning certificate directory...")
	}

	entries, err := os.ReadDir(listCertDir)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No certificate directory found.")
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
		certPath := filepath.Join(listCertDir, domainName, "cert.pem")

		if _, err := os.Stat(certPath); os.IsNotExist(err) {
			if verbose {
				log.Printf("Skipping %s: no cert.pem found", domainName)
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
