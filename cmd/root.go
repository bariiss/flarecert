package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "flarecert",
	Short: "Generate SSL certificates using Let's Encrypt with Cloudflare DNS-01 challenge",
	Long: `FlareCert is a CLI tool for generating and managing SSL certificates
from Let's Encrypt using Cloudflare's DNS-01 challenge method.

This tool is specifically designed to work with Cloudflare-proxied domains
(orange cloud enabled) and supports wildcard certificates.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Add global flags
	rootCmd.PersistentFlags().StringP("config", "c", "", "config file (default is .env)")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
}
