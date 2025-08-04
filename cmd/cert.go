package cmd

import (
	"fmt"
	"log"
	"strings"

	"github.com/bariiss/flarecert/internal/certificate"
	"github.com/bariiss/flarecert/internal/k8s"
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
  flarecert cert --domain example.com --domain www.example.com --domain api.example.com

  # Generate certificate with Kubernetes Secret YAML
  flarecert cert --domain example.com --k8s

  # Force renewal and create Kubernetes secret
  flarecert cert --domain example.com --force --k8s`,
	RunE: runCertCommand,
}

var (
	domains       []string
	certDir       string
	staging       bool
	keyType       string
	forceRenew    bool
	createK8sYaml bool
)

func init() {
	rootCmd.AddCommand(certCmd)

	certCmd.Flags().StringSliceVarP(&domains, "domain", "d", []string{}, "Domain name(s) for the certificate (required)")
	certCmd.Flags().StringVar(&certDir, "cert-dir", "./certs", "Directory to store certificates")
	certCmd.Flags().BoolVar(&staging, "staging", false, "Use Let's Encrypt staging environment")
	certCmd.Flags().StringVar(&keyType, "key-type", "rsa2048", "Key type: rsa2048, rsa4096, ec256, ec384")
	certCmd.Flags().BoolVar(&forceRenew, "force", false, "Force renewal even if certificate is valid")
	certCmd.Flags().BoolVar(&createK8sYaml, "k8s", false, "Generate Kubernetes Secret YAML file")

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

	// Create certificate manager
	manager, err := certificate.NewManager(certDir, keyType, staging, forceRenew, verbose)
	if err != nil {
		return fmt.Errorf("failed to create certificate manager: %w", err)
	}

	// Generate certificate
	if err := manager.GenerateCertificate(domains); err != nil {
		return err
	}

	// Create Kubernetes Secret YAML if requested
	if createK8sYaml {
		paths := utils.GetCertificatePathsForDomains(certDir, domains)

		// Determine primary domain (prefer wildcard for naming)
		primaryDomain := domains[0]
		for _, domain := range domains {
			if strings.HasPrefix(domain, "*.") {
				primaryDomain = domain
				break
			}
		}

		secretGen := k8s.NewSecretGenerator(verbose)
		if err := secretGen.CreateSecret(&paths, primaryDomain, domains); err != nil {
			log.Printf("Warning: failed to create Kubernetes secret YAML: %v", err)
		}
	}

	return nil
}
