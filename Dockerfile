# Build stage
FROM --platform=$BUILDPLATFORM golang:1.24-alpine AS builder

# Build arguments
ARG BUILDPLATFORM
ARG TARGETPLATFORM
ARG TARGETOS
ARG TARGETARCH
ARG VERSION=dev
ARG COMMIT=unknown
ARG DATE=unknown

# Install git and ca-certificates (needed for private repos and SSL)
RUN apk add --no-cache git ca-certificates tzdata

# Create appuser
RUN adduser -D -g '' appuser

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application with version information
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build \
    -ldflags="-w -s -extldflags '-static' -X github.com/bariiss/flarecert/cmd.version=${VERSION} -X github.com/bariiss/flarecert/cmd.commit=${COMMIT} -X github.com/bariiss/flarecert/cmd.date=${DATE}" \
    -a -installsuffix cgo \
    -o flarecert .

# Final stage
FROM scratch

# Import from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/passwd /etc/passwd

# Copy the binary
COPY --from=builder /build/flarecert /flarecert

# Use an unprivileged user
USER appuser

# Expose port (if needed for future web interface)
EXPOSE 8080

# Set entrypoint
ENTRYPOINT ["/flarecert"]

# Default command
CMD ["--help"]
