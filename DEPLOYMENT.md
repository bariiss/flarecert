# Deployment Guide for FlareCert

This guide covers various deployment scenarios for FlareCert in production environments.

## Table of Contents

1. [Server Deployment](#server-deployment)
2. [Docker Deployment](#docker-deployment)
3. [Automated Renewal](#automated-renewal)
4. [Security Best Practices](#security-best-practices)
5. [Integration Examples](#integration-examples)

## Server Deployment

### 1. Install FlareCert on Your Server

```bash
# Clone the repository
git clone <your-repo-url>
cd flarecert

# Build the binary
make build

# Install system-wide (optional)
sudo cp bin/flarecert /usr/local/bin/
```

### 2. Configuration

Create a production configuration file:

```bash
# Create secure directory
sudo mkdir -p /etc/flarecert
sudo chown root:root /etc/flarecert
sudo chmod 755 /etc/flarecert

# Create configuration
sudo tee /etc/flarecert/config.env << EOF
CLOUDFLARE_API_TOKEN=your_api_token_here
CLOUDFLARE_EMAIL=your_email@example.com
ACME_EMAIL=your_email@example.com
ACME_SERVER=https://acme-v02.api.letsencrypt.org/directory
CERT_DIR=/etc/ssl/flarecert
DNS_PROPAGATION_TIMEOUT=300
EOF

# Secure the config file
sudo chown root:root /etc/flarecert/config.env
sudo chmod 600 /etc/flarecert/config.env
```

### 3. Certificate Directory Setup

```bash
# Create certificate directory
sudo mkdir -p /etc/ssl/flarecert
sudo chown root:root /etc/ssl/flarecert
sudo chmod 755 /etc/ssl/flarecert
```

## Docker Deployment

### Dockerfile

```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o flarecert main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/flarecert .
COPY --from=builder /app/.env.example .

CMD ["./flarecert"]
```

### Docker Compose

```yaml
version: '3.8'
services:
  flarecert:
    build: .
    environment:
      - CLOUDFLARE_API_TOKEN=${CLOUDFLARE_API_TOKEN}
      - CLOUDFLARE_EMAIL=${CLOUDFLARE_EMAIL}
      - ACME_EMAIL=${ACME_EMAIL}
    volumes:
      - ./certs:/app/certs
    command: ["./flarecert", "cert", "--domain", "example.com"]
```

## Automated Renewal

### 1. Systemd Timer (Recommended)

Create a systemd service:

```bash
sudo tee /etc/systemd/system/flarecert-renew.service << EOF
[Unit]
Description=FlareCert Certificate Renewal
After=network.target

[Service]
Type=oneshot
User=root
EnvironmentFile=/etc/flarecert/config.env
ExecStart=/usr/local/bin/flarecert renew --cert-dir=/etc/ssl/flarecert
ExecStartPost=/bin/systemctl reload nginx
EOF
```

Create a systemd timer:

```bash
sudo tee /etc/systemd/system/flarecert-renew.timer << EOF
[Unit]
Description=Run FlareCert renewal daily
Requires=flarecert-renew.service

[Timer]
OnCalendar=daily
Persistent=true

[Install]
WantedBy=timers.target
EOF
```

Enable and start the timer:

```bash
sudo systemctl daemon-reload
sudo systemctl enable flarecert-renew.timer
sudo systemctl start flarecert-renew.timer
```

### 2. Cron Job Alternative

```bash
# Add to root's crontab
sudo crontab -e

# Add this line to run daily at 2 AM
0 2 * * * /usr/local/bin/flarecert renew --cert-dir=/etc/ssl/flarecert && systemctl reload nginx
```

## Security Best Practices

### 1. API Token Security

- Use Cloudflare API tokens with minimal permissions
- Scope tokens to specific zones only
- Rotate tokens regularly
- Store tokens in secure credential management systems

### 2. File Permissions

```bash
# Certificate files should be readable only by root and web server
sudo chown root:ssl-cert /etc/ssl/flarecert/*/cert.pem
sudo chown root:ssl-cert /etc/ssl/flarecert/*/fullchain.pem
sudo chmod 644 /etc/ssl/flarecert/*/cert.pem
sudo chmod 644 /etc/ssl/flarecert/*/fullchain.pem

# Private keys should be readable only by root and web server
sudo chown root:ssl-cert /etc/ssl/flarecert/*/privkey.pem
sudo chmod 640 /etc/ssl/flarecert/*/privkey.pem
```

### 3. Network Security

- Run FlareCert on secure networks only
- Use VPN for remote access
- Monitor certificate generation logs

## Integration Examples

### 1. Nginx Configuration

```nginx
server {
    listen 443 ssl http2;
    server_name example.com;

    ssl_certificate /etc/ssl/flarecert/example.com/fullchain.pem;
    ssl_certificate_key /etc/ssl/flarecert/example.com/privkey.pem;

    # Modern SSL configuration
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-RSA-AES256-GCM-SHA512:DHE-RSA-AES256-GCM-SHA512;
    ssl_prefer_server_ciphers off;

    # Security headers
    add_header Strict-Transport-Security "max-age=63072000" always;
    add_header X-Content-Type-Options nosniff;
    add_header X-Frame-Options DENY;

    location / {
        # Your application
    }
}
```

### 2. Apache Configuration

```apache
<VirtualHost *:443>
    ServerName example.com
    
    SSLEngine on
    SSLCertificateFile /etc/ssl/flarecert/example.com/cert.pem
    SSLCertificateKeyFile /etc/ssl/flarecert/example.com/privkey.pem
    SSLCertificateChainFile /etc/ssl/flarecert/example.com/chain.pem
    
    # Modern SSL configuration
    SSLProtocol all -SSLv3 -TLSv1 -TLSv1.1
    SSLCipherSuite ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384
    SSLHonorCipherOrder off
    
    # Security headers
    Header always set Strict-Transport-Security "max-age=63072000"
</VirtualHost>
```

### 3. Post-Renewal Hooks

Create a script for post-renewal actions:

```bash
#!/bin/bash
# /usr/local/bin/flarecert-post-renew.sh

set -e

echo "$(date): Certificate renewal completed"

# Reload web servers
if systemctl is-active --quiet nginx; then
    systemctl reload nginx
    echo "$(date): Nginx reloaded"
fi

if systemctl is-active --quiet apache2; then
    systemctl reload apache2
    echo "$(date): Apache reloaded"
fi

# Send notification (optional)
# curl -X POST "https://your-webhook-url" -d "FlareCert renewed certificates"

echo "$(date): Post-renewal hooks completed"
```

Make it executable and use in systemd service:

```bash
sudo chmod +x /usr/local/bin/flarecert-post-renew.sh

# Update systemd service
ExecStartPost=/usr/local/bin/flarecert-post-renew.sh
```

## Monitoring and Alerting

### 1. Certificate Expiration Monitoring

```bash
#!/bin/bash
# /usr/local/bin/check-cert-expiry.sh

CERT_DIR="/etc/ssl/flarecert"
WEBHOOK_URL="https://your-webhook-url"

for domain_dir in "$CERT_DIR"/*; do
    if [ -d "$domain_dir" ]; then
        domain=$(basename "$domain_dir")
        cert_file="$domain_dir/cert.pem"
        
        if [ -f "$cert_file" ]; then
            expiry=$(openssl x509 -enddate -noout -in "$cert_file" | cut -d= -f2)
            expiry_epoch=$(date -d "$expiry" +%s)
            current_epoch=$(date +%s)
            days_left=$(( (expiry_epoch - current_epoch) / 86400 ))
            
            if [ $days_left -lt 7 ]; then
                curl -X POST "$WEBHOOK_URL" -d "Certificate for $domain expires in $days_left days"
            fi
        fi
    fi
done
```

### 2. Log Monitoring

Monitor FlareCert logs for failures:

```bash
# Use journalctl to monitor systemd service logs
journalctl -u flarecert-renew.service -f

# Or monitor log files
tail -f /var/log/flarecert.log
```

## Backup and Recovery

### 1. Certificate Backup

```bash
#!/bin/bash
# Backup certificates
tar -czf "/backup/flarecert-$(date +%Y%m%d).tar.gz" /etc/ssl/flarecert/
```

### 2. Configuration Backup

```bash
# Backup configuration (excluding sensitive data)
cp /etc/flarecert/config.env.template /backup/
```

## Troubleshooting

### Common Issues

1. **DNS Propagation Timeout**: Increase `DNS_PROPAGATION_TIMEOUT`
2. **Rate Limits**: Use staging environment for testing
3. **Permissions**: Check file permissions and ownership
4. **API Token**: Verify token permissions and validity

### Debug Mode

Run with verbose output for troubleshooting:

```bash
flarecert cert --domain example.com --verbose --staging
```
