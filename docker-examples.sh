#!/bin/bash

# FlareCert Docker Examples
# This script demonstrates various ways to use FlareCert with Docker

set -e

echo "🐋 FlareCert Docker Examples"
echo "============================"

# Check if .env file exists
if [ ! -f .env ]; then
    echo "❌ .env file not found. Please create one from .env.example"
    echo "   cp .env.example .env"
    echo "   # Then edit .env with your Cloudflare credentials"
    exit 1
fi

echo "✅ Environment file found"
echo ""

echo "📋 Available Docker commands:"
echo ""

echo "1. 🔨 Build the Docker image:"
echo "   docker build -t flarecert ."
echo ""

echo "2. 🌐 List your Cloudflare zones:"
echo "   docker run --rm --env-file .env flarecert zones"
echo ""

echo "3. 🔐 Generate a test certificate (staging):"
echo "   docker run --rm --env-file .env -v \$(pwd)/certs:/certs flarecert cert --domain test.example.com --staging --verbose"
echo ""

echo "4. 🌟 Generate a wildcard certificate:"
echo "   docker run --rm --env-file .env -v \$(pwd)/certs:/certs flarecert cert --domain \"*.example.com\" --staging"
echo ""

echo "5. 📋 List generated certificates:"
echo "   docker run --rm --env-file .env -v \$(pwd)/certs:/certs flarecert list"
echo ""

echo "6. 🔄 Renew certificates:"
echo "   docker run --rm --env-file .env -v \$(pwd)/certs:/certs flarecert renew --verbose"
echo ""

echo "7. 🚀 Using Docker Compose:"
echo "   # Generate certificate"
echo "   docker-compose run --rm flarecert cert --domain example.com --staging"
echo ""
echo "   # Renew certificates"
echo "   docker-compose --profile renew up"
echo ""
echo "   # Start with nginx (after setting up certificates)"
echo "   docker-compose --profile nginx up -d"
echo ""

echo "8. 📅 Scheduled renewals with cron:"
echo "   # Add to crontab:"
echo "   0 2 * * * cd /path/to/flarecert && docker-compose --profile renew up --abort-on-container-exit"
echo ""

echo "💡 Tips:"
echo "  - Always test with --staging first"
echo "  - Mount the certs directory to persist certificates"
echo "  - Use --env-file to load environment variables securely"
echo "  - Check certificate files in ./certs/ after generation"
echo ""

echo "🔧 Quick start:"
echo "  1. Build: docker build -t flarecert ."
echo "  2. Test:  docker run --rm --env-file .env flarecert zones"
echo "  3. Cert:  docker run --rm --env-file .env -v \$(pwd)/certs:/certs flarecert cert --domain yourdomain.com --staging"
