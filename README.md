# FlareCert - Let's Encrypt SSL Certificates with Cloudflare DNS-01 Challenge

A Go application for automatically generating and renewing SSL certificates from Let's Encrypt using Cloudflare's DNS-01 challenge method.

## Features

- ✅ Automatic SSL certificate generation using Let's Encrypt
- ✅ DNS-01 challenge via Cloudflare API
- ✅ Interactive zone selection for multi-zone accounts
- ✅ Support for wildcard certificates with smart directory naming
- ✅ Works with Cloudflare's proxy (orange cloud) enabled
- ✅ Automatic certificate renewal
- ✅ Multiple domain support (SAN certificates)
- ✅ Organized certificate storage with archiving
- ✅ Certificate metadata tracking
- ✅ Kubernetes Secret YAML generation (clean, minimal format)
- ✅ Export existing certificates to Kubernetes Secrets
- ✅ Shell completion with domain suggestions
- ✅ Secure credential management

## Prerequisites

1. Cloudflare account with domains managed
2. Cloudflare API token with DNS edit permissions
3. Go 1.24 or later

## Installation

### Option 1: Homebrew (macOS)

Install FlareCert using Homebrew:

```bash
brew tap bariiss/flarecert
brew install flarecert
```

For updates:
```bash
brew update
brew upgrade flarecert
```

### Option 2: Install using Go (Recommended for other platforms)

```bash
go install github.com/bariiss/flarecert@latest
```

### Option 3: Download pre-built binaries

Download the appropriate binary for your system from the [releases page](https://github.com/bariiss/flarecert/releases):

- **Linux (x64)**: `flarecert-linux-amd64`
- **Linux (ARM64)**: `flarecert-linux-arm64`
- **macOS (Intel)**: `flarecert-darwin-amd64`
- **macOS (Apple Silicon)**: `flarecert-darwin-arm64`
- **Windows (x64)**: `flarecert-windows-amd64.exe`

Make the binary executable and move it to your PATH:
```bash
chmod +x flarecert-*
sudo mv flarecert-* /usr/local/bin/flarecert
```

### Option 4: Build from source

1. Clone the repository:
```bash
git clone https://github.com/bariiss/flarecert.git
cd flarecert
```

2. Install dependencies:
```bash
go mod tidy
```

3. Build the binary:
```bash
go build -o flarecert main.go
```

## Setup

### Environment Configuration

FlareCert supports multiple ways to configure your credentials with the following precedence (highest to lowest):

1. **Local `.env` file** (highest priority)
2. **System environment variables** 
3. **Default values** (if any)

#### Option 1: Using .env file (Recommended for development)

1. Create a `.env` file:
```bash
cp .env.example .env
```

2. Configure your Cloudflare credentials in `.env`:
```
CLOUDFLARE_API_TOKEN=your_api_token_here
CLOUDFLARE_EMAIL=your_email@example.com
ACME_EMAIL=your_email@example.com
```

#### Option 2: Using system environment variables (Recommended for production)

Export the variables in your shell or add them to your system's environment:

```bash
export CLOUDFLARE_API_TOKEN=your_api_token_here
export CLOUDFLARE_EMAIL=your_email@example.com
export ACME_EMAIL=your_email@example.com
```

**Note:** If both `.env` file and system environment variables are present, the `.env` file values will take precedence.

## Usage

### List available Cloudflare zones:
```bash
flarecert zones
```

### Generate a certificate for a single domain:
```bash
flarecert cert --domain example.com
```

### Generate a wildcard certificate:
```bash
flarecert cert --domain "*.example.com"
```

### Generate a certificate for multiple domains (wildcard + apex):
```bash
flarecert cert --domain example.com --domain "*.example.com"
```

### Generate a certificate for multiple domains:
```bash
flarecert cert --domain example.com --domain www.example.com --domain api.example.com
```

### Generate a certificate with Kubernetes Secret YAML:
```bash
flarecert cert --domain example.com --k8s
```

### List existing certificates:
```bash
flarecert list
```

### Renew existing certificates:
```bash
flarecert renew
```

### Export existing certificates to Kubernetes Secrets:
```bash
# List available certificates for export
flarecert export

# Export specific certificate
flarecert export --domain example.com

# Export all certificates
flarecert export --all

# Export to custom directory
flarecert export --domain example.com --output ./k8s-secrets/
```

## Shell Completion

FlareCert supports auto-completion for bash, zsh, fish, and PowerShell shells. This provides tab-completion for commands, flags, and even domain suggestions from your Cloudflare zones.

### Setup Completion

#### For Zsh (macOS default):
```bash
# Generate completion and save to zsh completion directory
flarecert completion zsh > "${fpath[1]}/_flarecert"

# Or for Homebrew users:
flarecert completion zsh > $(brew --prefix)/share/zsh/site-functions/_flarecert

# Reload your shell or run:
source ~/.zshrc
```

#### For Bash:
```bash
# For current session:
source <(flarecert completion bash)

# For all sessions (Linux):
flarecert completion bash > /etc/bash_completion.d/flarecert

# For all sessions (macOS with Homebrew):
flarecert completion bash > $(brew --prefix)/etc/bash_completion.d/flarecert
```

#### For Fish:
```bash
# For current session:
flarecert completion fish | source

# For all sessions:
flarecert completion fish > ~/.config/fish/completions/flarecert.fish
```

### Completion Features

- ✅ Command and flag completion
- ✅ Domain suggestions from your Cloudflare zones
- ✅ Key type options (rsa2048, rsa4096, ec256, ec384)
- ✅ Wildcard domain suggestions (*.domain.com)
- ✅ Common subdomain suggestions (www.domain.com)
- ✅ Export command domain completion
- ✅ Certificate directory completion

### Example Usage with Completion:
```bash
# Type and press TAB to see domain suggestions from your zones:
flarecert cert --domain <TAB>

# See available key types:
flarecert cert --domain example.com --key-type <TAB>

# Export with domain completion:
flarecert export --domain <TAB>
```

## Command Reference

### Available Commands

| Command | Description |
|---------|-------------|
| `flarecert zones` | List all Cloudflare zones in your account |
| `flarecert cert` | Generate new SSL certificates |
| `flarecert list` | List existing certificates |
| `flarecert renew` | Renew existing certificates |
| `flarecert export` | Export existing certificates to Kubernetes Secrets |
| `flarecert completion` | Generate shell completion scripts |
| `flarecert version` | Show version information |

### Certificate Generation Options

| Flag | Description | Example |
|------|-------------|---------|
| `--domain` | Domain name(s) to generate certificate for | `--domain example.com` |
| `--key-type` | Certificate key type (rsa2048, rsa4096, ec256, ec384) | `--key-type rsa4096` |
| `--staging` | Use Let's Encrypt staging environment for testing | `--staging` |
| `--force` | Force renewal without prompting | `--force` |
| `--k8s` | Generate Kubernetes Secret YAML | `--k8s` |
| `--cert-dir` | Custom certificate storage directory | `--cert-dir ./my-certs` |

### Export Options

| Flag | Description | Example |
|------|-------------|---------|
| `--domain` | Specific domain to export | `--domain example.com` |
| `--all` | Export all available certificates | `--all` |
| `--output` | Custom output directory for YAML files | `--output ./k8s-secrets/` |
| `--cert-dir` | Certificate directory to read from | `--cert-dir ./certs` |

### Global Options

| Flag | Description |
|------|-------------|
| `-v, --verbose` | Enable verbose output |
| `-c, --config` | Config file path (default: .env) |

## ACME Challenge Methods

### Why DNS-01 is preferred for Cloudflare:

1. **HTTP-01 Challenge**: 
   - Requires serving files at `http://<domain>/.well-known/acme-challenge/`
   - ❌ Doesn't work with Cloudflare's proxy (orange cloud)
   - ❌ Cannot generate wildcard certificates

2. **TLS-ALPN-01 Challenge**:
   - Uses TLS negotiation on port 443
   - ❌ Conflicts with Cloudflare's TLS termination
   - ❌ Cannot generate wildcard certificates

3. **DNS-01 Challenge**:
   - ✅ Creates TXT records at `_acme-challenge.<domain>`
   - ✅ Works perfectly with Cloudflare's proxy
   - ✅ Supports wildcard certificates
   - ✅ Can be fully automated via API
   - ✅ No server downtime required

## Certificate Storage

Certificates are stored in an organized structure in the `certs/` directory with smart naming:

### Directory Naming Convention
- **Regular domains**: `example.com/`
- **Wildcard certificates**: `wildcard-example-com/` (prioritized when both apex and wildcard domains are requested)
- **Mixed certificates**: When requesting both `example.com` and `*.example.com`, the directory will be named `wildcard-example-com/`

### Directory Structure
```
certs/
├── example.com/
│   ├── current/          # Active certificate files
│   │   ├── cert.pem      # Certificate
│   │   ├── privkey.pem   # Private key
│   │   ├── chain.pem     # Certificate chain
│   │   ├── fullchain.pem # Full certificate chain
│   │   ├── cert.json     # Certificate metadata
│   │   └── example-com-tls-secret.yaml  # Kubernetes Secret (if --k8s flag used)
│   ├── archive/          # Previous certificates
│   │   └── cert-20240801-120000-*.pem
│   └── logs/             # Certificate generation logs
└── wildcard-example-com/ # Wildcard certificates
    ├── current/
    │   ├── cert.pem
    │   ├── privkey.pem
    │   ├── chain.pem
    │   ├── fullchain.pem
    │   ├── cert.json
    │   └── wildcard-example-com-tls-secret.yaml
    ├── archive/
    └── logs/
```

### Kubernetes Secret Generation

The generated Kubernetes Secret YAML files are clean and minimal:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: wildcard-example-com-tls
  namespace: default
type: kubernetes.io/tls
data:
  tls.crt: <base64-encoded-certificate>
  tls.key: <base64-encoded-private-key>
  ca.crt: <base64-encoded-certificate-chain>
```

Apply to your cluster:
```bash
kubectl apply -f wildcard-example-com-tls-secret.yaml
```

## Security Notes

- Store API tokens securely
- Use environment variables for production deployments
- Use `.env` files for local development only
- Never commit `.env` files to version control
- Regularly rotate API tokens
- Monitor certificate expiration dates

## License

MIT License
