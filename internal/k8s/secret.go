package k8s

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bariiss/flarecert/internal/utils"
)

// SecretGenerator handles Kubernetes Secret generation
type SecretGenerator struct {
	verbose bool
}

// NewSecretGenerator creates a new Kubernetes secret generator
func NewSecretGenerator(verbose bool) *SecretGenerator {
	return &SecretGenerator{
		verbose: verbose,
	}
}

// CreateSecret creates a Kubernetes Secret YAML file for the certificate
func (sg *SecretGenerator) CreateSecret(paths *utils.CertificatePaths, primaryDomain string, domains []string) error {
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
	secretName := GenerateSecretName(primaryDomain)

	// Create YAML content
	yamlContent := sg.generateYAMLContent(secretName, primaryDomain, domains, certB64, keyB64, fullchainB64)

	// Write YAML file
	yamlFile := filepath.Join(paths.CurrentDir, fmt.Sprintf("%s-secret.yaml", secretName))
	if err := os.WriteFile(yamlFile, []byte(yamlContent), 0644); err != nil {
		return fmt.Errorf("failed to write Kubernetes secret YAML: %w", err)
	}

	fmt.Printf("🚀 Kubernetes Secret YAML created: %s\n", yamlFile)
	fmt.Printf("📝 Usage: kubectl apply -f %s\n", yamlFile)

	return nil
}

// generateYAMLContent generates the Kubernetes Secret YAML content
func (sg *SecretGenerator) generateYAMLContent(secretName, primaryDomain string, domains []string, certB64, keyB64, fullchainB64 string) string {
	return fmt.Sprintf(`apiVersion: v1
kind: Secret
metadata:
  name: %s
  namespace: default
type: kubernetes.io/tls
data:
  tls.crt: %s
  tls.key: %s
  ca.crt: %s
`,
		secretName,
		certB64,
		keyB64,
		fullchainB64,
	)
}

// GenerateSecretName generates a Kubernetes-safe secret name
func GenerateSecretName(domain string) string {
	// Remove wildcard prefix and replace dots with hyphens
	name := strings.Replace(domain, "*.", "wildcard-", 1)
	name = strings.Replace(name, ".", "-", -1)

	// Ensure it starts and ends with alphanumeric characters
	name = strings.Trim(name, "-")

	// Add suffix to avoid conflicts
	return fmt.Sprintf("%s-tls", name)
}
