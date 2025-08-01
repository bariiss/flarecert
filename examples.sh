#!/bin/bash

# FlareCert - Example Usage Script
# This script demonstrates various ways to use FlareCert

set -e

echo "🚀 FlareCert - SSL Certificate Generation Examples"
echo "=================================================="

# Check if .env file exists
if [ ! -f .env ]; then
    echo "❌ .env file not found. Please create one from .env.example"
    echo "   cp .env.example .env"
    echo "   # Then edit .env with your Cloudflare credentials"
    exit 1
fi

# Source environment variables
source .env

# Check required environment variables
if [ -z "$CLOUDFLARE_API_TOKEN" ] || [ -z "$CLOUDFLARE_EMAIL" ] || [ -z "$ACME_EMAIL" ]; then
    echo "❌ Missing required environment variables:"
    echo "   CLOUDFLARE_API_TOKEN, CLOUDFLARE_EMAIL, ACME_EMAIL"
    echo "   Please check your .env file"
    exit 1
fi

echo "✅ Environment variables loaded"
echo ""

# Example 0: List available zones
echo "📋 Example 0: List available Cloudflare zones"
echo "Command: ./bin/flarecert zones --verbose"
echo "This shows all zones in your Cloudflare account"
echo ""

# Example 1: Single domain certificate (staging)
echo "📋 Example 1: Single domain certificate (staging)"
echo "Command: ./bin/flarecert cert --domain test.example.com --staging --verbose"
echo "This generates a staging certificate for a single domain"
echo ""

# Example 2: Wildcard certificate (staging)
echo "📋 Example 2: Wildcard certificate (staging)"
echo "Command: ./bin/flarecert cert --domain \"*.example.com\" --staging --verbose"
echo "This generates a staging wildcard certificate"
echo ""

# Example 3: Multi-domain certificate (SAN)
echo "📋 Example 3: Multi-domain certificate (SAN)"
echo "Command: ./bin/flarecert cert --domain example.com --domain www.example.com --domain api.example.com --staging --verbose"
echo "This generates a staging certificate for multiple domains"
echo ""

# Example 4: Production certificate
echo "📋 Example 4: Production certificate"
echo "Command: ./bin/flarecert cert --domain example.com --verbose"
echo "This generates a production certificate (be careful!)"
echo ""

# Example 5: List certificates
echo "📋 Example 5: List certificates"
echo "Command: ./bin/flarecert list --verbose"
echo "This lists all stored certificates with their status"
echo ""

# Example 6: Renew certificates
echo "📋 Example 6: Renew certificates"
echo "Command: ./bin/flarecert renew --verbose"
echo "This renews certificates that expire within 30 days"
echo ""

echo "⚠️  Important Notes:"
echo "  - Always test with --staging first"
echo "  - Let's Encrypt has rate limits for production"
echo "  - DNS-01 challenge works with Cloudflare proxy enabled"
echo "  - Certificates are stored in ./certs/ directory"
echo ""

echo "🔧 To run a test certificate:"
echo "  # First, check your zones:"
echo "  ./bin/flarecert zones"
echo ""
echo "  # Then generate a test certificate:"
echo "  ./bin/flarecert cert --domain test.yourdomain.com --staging --verbose"
