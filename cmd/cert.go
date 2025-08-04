package cmd

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"path/filepath"
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
	client, err := acme.NewClient(cfg, verbose, keyType)
	if err != nil {
		return fmt.Errorf("failed to create ACME client: %w", err)
	}

	// Check if certificate already exists and handle accordingly
	if !forceRenew {
		// Check if certificate exists and get its info
		if _, err := os.Stat(paths.CertFile); err == nil {
			// Certificate file exists, parse it to get expiration info
			existingDomains, expiresAt, err := acme.ParseCertificateInfo(paths.CertFile)
			if err != nil {
				if verbose {
					log.Printf("Warning: failed to parse existing certificate: %v", err)
				}
			} else {
				// Check if domains match
				if domainsMatch(domains, existingDomains) {
					now := time.Now()
					daysRemaining := int(expiresAt.Sub(now).Hours() / 24)

					if expiresAt.Before(now) {
						// Certificate is expired, renew automatically
						fmt.Printf("⚠️  Certificate for %s has expired (%d days ago)\n",
							strings.Join(domains, ", "), -daysRemaining)
						fmt.Printf("🔄 Automatically renewing expired certificate...\n")
					} else if daysRemaining <= 30 {
						// Certificate expires soon, ask user
						fmt.Printf("⚠️  Certificate for %s expires in %d days (%s)\n",
							strings.Join(domains, ", "), daysRemaining, expiresAt.Format("2006-01-02 15:04"))

						if !askUserConfirmation("Do you want to renew it now?") {
							fmt.Printf("Certificate renewal cancelled.\n")
							return nil
						}
						fmt.Printf("🔄 Renewing certificate...\n")
					} else {
						// Certificate is valid and has plenty of time left
						fmt.Printf("✅ Certificate for %s is already valid and expires in %d days (%s)\n",
							strings.Join(domains, ", "), daysRemaining, expiresAt.Format("2006-01-02 15:04"))

						if !askUserConfirmation("Do you want to renew it anyway?") {
							fmt.Printf("Certificate generation cancelled. Use --force to renew without prompting.\n")
							return nil
						}
						fmt.Printf("🔄 Force renewing certificate...\n")
					}
				} else {
					// Different domains, ask user what to do
					fmt.Printf("⚠️  Found existing certificate with different domains:\n")
					fmt.Printf("   Existing: %s\n", strings.Join(existingDomains, ", "))
					fmt.Printf("   Requested: %s\n", strings.Join(domains, ", "))

					if !askUserConfirmation("Do you want to replace it with the new certificate?") {
						fmt.Printf("Certificate generation cancelled.\n")
						return nil
					}
					fmt.Printf("🔄 Replacing certificate with new domains...\n")
				}
			}
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

	cert, err := client.ObtainCertificate(domains)
	if err != nil {
		return fmt.Errorf("failed to obtain certificate: %w", err)
	}

	// Save certificate files with improved structure
	// Trim leading/trailing whitespace from IssuerCertificate to avoid empty lines
	trimmedIssuerCert := []byte(strings.TrimSpace(string(cert.IssuerCertificate)))

	files := map[string][]byte{
		paths.CertFile:      cert.Certificate,
		paths.KeyFile:       cert.PrivateKey,
		paths.ChainFile:     trimmedIssuerCert,
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

	// Create Kubernetes Secret YAML if requested
	if createK8sYaml {
		if err := createKubernetesSecret(&paths, primaryDomain, domains); err != nil {
			log.Printf("Warning: failed to create Kubernetes secret YAML: %v", err)
		}
	}

	fmt.Printf("✅ Certificate successfully generated and saved to: %s\n", paths.CurrentDir)
	fmt.Printf("📅 Certificate expires: %s\n", cert.NotAfter.Format("2006-01-02 15:04:05 MST"))

	return nil
}

// askUserConfirmation prompts the user for yes/no confirmation
func askUserConfirmation(message string) bool {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("%s [y/N]: ", message)
		response, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("Error reading input: %v", err)
			return false
		}

		response = strings.ToLower(strings.TrimSpace(response))
		switch response {
		case "y", "yes":
			return true
		case "n", "no", "":
			return false
		default:
			fmt.Printf("Please answer yes (y) or no (n).\n")
		}
	}
}

// domainsMatch checks if two domain slices contain the same domains (ignoring duplicates)
func domainsMatch(domains1, domains2 []string) bool {
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

// createKubernetesSecret creates a Kubernetes Secret YAML file for the certificate
func createKubernetesSecret(paths *utils.CertificatePaths, primaryDomain string, domains []string) error {
	// Read certificate files
	certData, err := os.ReadFile(paths.CertFile)
	if err != nil {
		return fmt.Errorf("failed to read certificate file: %w", err)
	}

	keyData, err := os.ReadFile(paths.KeyFile)
	if err != nil {
		return fmt.Errorf("failed to read key file: %w", err)
	}

	fullchainData, err := os.ReadFile(paths.FullchainFile)
	if err != nil {
		return fmt.Errorf("failed to read fullchain file: %w", err)
	}

	// Encode data to base64
	certB64 := base64.StdEncoding.EncodeToString(certData)
	keyB64 := base64.StdEncoding.EncodeToString(keyData)
	fullchainB64 := base64.StdEncoding.EncodeToString(fullchainData)

	// Generate secret name (safe for Kubernetes)
	secretName := generateK8sSecretName(primaryDomain)

	// Create YAML content
	yamlContent := fmt.Sprintf(`apiVersion: v1
kind: Secret
metadata:
  name: %s
  namespace: default
  labels:
    app: flarecert
    domain: %s
    type: tls-certificate
  annotations:
    flarecert.io/domains: "%s"
    flarecert.io/primary-domain: "%s"
    flarecert.io/created-at: "%s"
    cert-manager.io/issuer-name: "letsencrypt"
type: kubernetes.io/tls
data:
  tls.crt: %s
  tls.key: %s
  ca.crt: %s
---
# Usage examples:
#
# 1. Apply this secret to your cluster:
#    kubectl apply -f %s-secret.yaml
#
# 2. Use in an Ingress:
#    apiVersion: networking.k8s.io/v1
#    kind: Ingress
#    metadata:
#      name: %s-ingress
#    spec:
#      tls:
#        - hosts:
#            - %s
#          secretName: %s
#      rules:
#        - host: %s
#          http:
#            paths:
#              - path: /
#                pathType: Prefix
#                backend:
#                  service:
#                    name: your-service
#                    port:
#                      number: 80
#
# 3. Use with cert-manager for auto-renewal:
#    Add this annotation to your Ingress:
#    cert-manager.io/cluster-issuer: "letsencrypt-prod"
`,
		secretName,
		primaryDomain,
		strings.Join(domains, ", "),
		primaryDomain,
		time.Now().Format(time.RFC3339),
		certB64,
		keyB64,
		fullchainB64,
		secretName,
		secretName,
		primaryDomain,
		secretName,
		primaryDomain,
	)

	// Write YAML file
	yamlFile := filepath.Join(paths.CurrentDir, fmt.Sprintf("%s-secret.yaml", secretName))
	if err := os.WriteFile(yamlFile, []byte(yamlContent), 0644); err != nil {
		return fmt.Errorf("failed to write Kubernetes secret YAML: %w", err)
	}

	fmt.Printf("🚀 Kubernetes Secret YAML created: %s\n", yamlFile)
	fmt.Printf("📝 Usage: kubectl apply -f %s\n", yamlFile)

	return nil
}

// generateK8sSecretName generates a Kubernetes-safe secret name
func generateK8sSecretName(domain string) string {
	// Remove wildcard prefix and replace dots with hyphens
	name := strings.Replace(domain, "*.", "wildcard-", 1)
	name = strings.Replace(name, ".", "-", -1)

	// Ensure it starts and ends with alphanumeric characters
	name = strings.Trim(name, "-")

	// Add suffix to avoid conflicts
	return fmt.Sprintf("%s-tls", name)
}
