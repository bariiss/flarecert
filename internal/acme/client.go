package acme

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/bariiss/flarecert/internal/config"
	"github.com/bariiss/flarecert/internal/dns"

	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/registration"
)

// Client wraps the ACME client with our configuration
type Client struct {
	client  *lego.Client
	config  *config.Config
	verbose bool
}

// CertificateResult holds the certificate data
type CertificateResult struct {
	Certificate       []byte
	PrivateKey        []byte
	IssuerCertificate []byte
	NotAfter          time.Time
}

// User represents the ACME user
type User struct {
	Email        string
	Registration *registration.Resource
	key          *rsa.PrivateKey
}

// GetEmail returns the user's email
func (u *User) GetEmail() string {
	return u.Email
}

// GetRegistration returns the user's registration
func (u *User) GetRegistration() *registration.Resource {
	return u.Registration
}

// GetPrivateKey returns the user's private key
func (u *User) GetPrivateKey() crypto.PrivateKey {
	return u.key
}

// NewClient creates a new ACME client with Cloudflare DNS provider
func NewClient(cfg *config.Config, verbose bool, keyType string) (*Client, error) {
	// Generate user private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	// Create user
	user := &User{
		Email: cfg.ACMEEmail,
		key:   privateKey,
	}

	// Create lego config
	legoConfig := lego.NewConfig(user)
	legoConfig.CADirURL = cfg.ACMEServer

	// Set key type for certificates
	switch keyType {
	case "rsa2048":
		legoConfig.Certificate.KeyType = certcrypto.RSA2048
	case "rsa4096":
		legoConfig.Certificate.KeyType = certcrypto.RSA4096
	case "ec256":
		legoConfig.Certificate.KeyType = certcrypto.EC256
	case "ec384":
		legoConfig.Certificate.KeyType = certcrypto.EC384
	default:
		legoConfig.Certificate.KeyType = certcrypto.RSA2048
	}

	// Create lego client
	client, err := lego.NewClient(legoConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create lego client: %w", err)
	}

	// Create Cloudflare DNS provider
	provider, err := dns.NewCloudflareProvider(cfg.CloudflareAPIToken, cfg.CloudflareEmail, cfg.DNSTimeout, verbose)
	if err != nil {
		return nil, fmt.Errorf("failed to create Cloudflare provider: %w", err)
	}

	// Set DNS challenge provider
	err = client.Challenge.SetDNS01Provider(provider)
	if err != nil {
		return nil, fmt.Errorf("failed to set DNS provider: %w", err)
	}

	// Register user
	reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		return nil, fmt.Errorf("failed to register user: %w", err)
	}
	user.Registration = reg

	if verbose {
		log.Printf("ACME client registered with server: %s", cfg.ACMEServer)
	}

	return &Client{
		client:  client,
		config:  cfg,
		verbose: verbose,
	}, nil
}

// ObtainCertificate requests a new certificate for the given domains
func (c *Client) ObtainCertificate(domains []string) (*CertificateResult, error) {
	if c.verbose {
		log.Printf("Requesting certificate for domains: %v", domains)
	}

	// Create certificate request
	request := certificate.ObtainRequest{
		Domains: domains,
		Bundle:  true,
	}

	// Obtain certificate
	certificates, err := c.client.Certificate.Obtain(request)
	if err != nil {
		return nil, fmt.Errorf("failed to obtain certificate: %w", err)
	}

	// Parse certificate to get expiration
	block, _ := pem.Decode(certificates.Certificate)
	if block == nil {
		return nil, fmt.Errorf("failed to parse certificate PEM")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	if c.verbose {
		log.Printf("Certificate obtained successfully, expires: %s", cert.NotAfter.Format("2006-01-02 15:04:05"))
	}

	return &CertificateResult{
		Certificate:       certificates.Certificate,
		PrivateKey:        certificates.PrivateKey,
		IssuerCertificate: certificates.IssuerCertificate,
		NotAfter:          cert.NotAfter,
	}, nil
}

// IsCertificateValid checks if a certificate exists and is valid for the given domains
func IsCertificateValid(certPath string, domains []string) (bool, error) {
	// Check if certificate file exists
	certData, err := os.ReadFile(certPath)
	if err != nil {
		return false, err
	}

	// Parse certificate
	block, _ := pem.Decode(certData)
	if block == nil {
		return false, fmt.Errorf("failed to parse certificate PEM")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return false, fmt.Errorf("failed to parse certificate: %w", err)
	}

	// Check expiration (consider valid if expires more than 30 days from now)
	if time.Now().Add(30 * 24 * time.Hour).After(cert.NotAfter) {
		return false, nil // Certificate expires within 30 days
	}

	// Check if all requested domains are covered
	certDomains := make(map[string]bool)
	if cert.Subject.CommonName != "" {
		certDomains[cert.Subject.CommonName] = true
	}
	for _, domain := range cert.DNSNames {
		certDomains[domain] = true
	}

	for _, domain := range domains {
		if !certDomains[domain] {
			return false, nil // Domain not covered by certificate
		}
	}

	return true, nil
}

// ParseCertificateInfo extracts domain names and expiration from a certificate file
func ParseCertificateInfo(certPath string) ([]string, time.Time, error) {
	certData, err := os.ReadFile(certPath)
	if err != nil {
		return nil, time.Time{}, err
	}

	block, _ := pem.Decode(certData)
	if block == nil {
		return nil, time.Time{}, fmt.Errorf("failed to parse certificate PEM")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, time.Time{}, fmt.Errorf("failed to parse certificate: %w", err)
	}

	// Use a map to track which domains we've seen
	seen := make(map[string]bool)
	domains := make([]string, 0)

	// Add CommonName first if it exists and hasn't been seen
	if cert.Subject.CommonName != "" && !seen[cert.Subject.CommonName] {
		domains = append(domains, cert.Subject.CommonName)
		seen[cert.Subject.CommonName] = true
	}

	// Add DNS names that haven't been seen yet
	for _, dnsName := range cert.DNSNames {
		if !seen[dnsName] {
			domains = append(domains, dnsName)
			seen[dnsName] = true
		}
	}

	return domains, cert.NotAfter, nil
}
