package cmd

import (
	"fmt"
	"log"
	"os"
	"text/tabwriter"

	"github.com/bariiss/flarecert/internal/config"
	"github.com/bariiss/flarecert/internal/dns"
	"github.com/spf13/cobra"
)

var zonesCmd = &cobra.Command{
	Use:   "zones",
	Short: "List available Cloudflare zones",
	Long: `List all available Cloudflare zones in your account.

This command shows all zones that you can use for certificate generation,
along with their status and other information.`,
	RunE: runZonesCommand,
}

func init() {
	rootCmd.AddCommand(zonesCmd)
}

func runZonesCommand(cmd *cobra.Command, args []string) error {
	verbose, _ := cmd.Flags().GetBool("verbose")

	if verbose {
		log.Println("🔍 Fetching Cloudflare zones...")
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Create Cloudflare DNS provider
	provider, err := dns.NewCloudflareProvider(cfg.CloudflareAPIToken, cfg.CloudflareEmail, cfg.DNSTimeout, verbose)
	if err != nil {
		return fmt.Errorf("failed to create Cloudflare provider: %w", err)
	}

	// List zones
	zones, err := provider.ListZones()
	if err != nil {
		return fmt.Errorf("failed to list zones: %w", err)
	}

	if len(zones) == 0 {
		fmt.Println("❌ No zones found in your Cloudflare account")
		return nil
	}

	// Create table writer
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	defer w.Flush()

	fmt.Printf("📋 Found %d zone(s) in your Cloudflare account:\n\n", len(zones))
	fmt.Fprintln(w, "STATUS\tZONE NAME\tZONE ID")
	fmt.Fprintln(w, "------\t---------\t-------")

	for _, zone := range zones {
		status := "✅ Active"
		if zone.Status != "active" {
			status = fmt.Sprintf("⚠️  %s", zone.Status)
		}

		fmt.Fprintf(w, "%s\t%s\t%s\n", status, zone.Name, zone.ID)
	}

	fmt.Println("\n💡 Tips:")
	fmt.Println("  - Only active zones can be used for certificate generation")
	fmt.Println("  - FlareCert will automatically detect the right zone for your domain")
	fmt.Println("  - If multiple zones match, you'll be prompted to choose")

	return nil
}
