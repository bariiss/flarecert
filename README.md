# FlareCert - Let's Encrypt SSL Certificates with Cloudflare DNS-01 Challenge

A Go application for automatically generating and renewing SSL certificates from Let's Encrypt using Cloudflare's DNS-01 challenge method.

## Features

- ✅ Automatic SSL certificate generation using Let's Encrypt
- ✅ DNS-01 challenge via Cloudflare API
- ✅ Interactive zone selection for multi-zone accounts
- ✅ Support for wildcard certificates
- ✅ Works with Cloudflare's proxy (orange cloud) enabled
- ✅ Automatic certificate renewal
- ✅ Multiple domain support (SAN certificates)
- ✅ Organized certificate storage with archiving
- ✅ Certificate metadata tracking
- ✅ Secure credential management

## Prerequisites

1. Cloudflare account with domains managed
2. Cloudflare API token with DNS edit permissions
3. Go 1.24 or later

## Installation

### Option 1: Install using Go (Recommended)

```bash
go install github.com/bariiss/flarecert@latest
```

### Option 2: Download pre-built binaries

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

### Option 3: Build from source

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

### Generate a certificate for multiple domains:
```bash
flarecert cert --domain example.com --domain www.example.com --domain api.example.com
```

### List existing certificates:
```bash
flarecert list
```

### Renew existing certificates:
```bash
flarecert renew
```

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

Certificates are stored in an organized structure in the `certs/` directory:
```
certs/
├── example.com/
│   ├── current/          # Active certificate files
│   │   ├── cert.pem      # Certificate
│   │   ├── privkey.pem   # Private key
│   │   ├── chain.pem     # Certificate chain
│   │   ├── fullchain.pem # Full certificate chain
│   │   └── cert.json     # Certificate metadata
│   ├── archive/          # Previous certificates
│   │   └── cert-20240801-120000-*.pem
│   └── logs/             # Certificate generation logs
└── wildcard.example.com/ # Wildcard certificates
    └── ...
```

## Security Notes

- Store API tokens securely
- Use environment variables for production
- Regularly rotate API tokens
- Monitor certificate expiration dates

## License

MIT License
