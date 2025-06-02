# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /workspace

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY cmd/ cmd/
COPY api/ api/
COPY controllers/ controllers/
COPY internal/ internal/

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags '-extldflags "-static"' -o manager cmd/main.go

# Final stage - using distroless for minimal attack surface
FROM docker.io/alpine:3.22

RUN apk add --no-cache \
    git==2.49.0-r0 \
    helm==3.18.0-r0 \
    ca-certificates==20241121-r2 \
    && addgroup -g 65532 -S nonroot \
    && adduser -u 65532 -S nonroot -G nonroot

WORKDIR /

# Copy the binary from builder
COPY --from=builder /workspace/manager .

# Run as non-root user
USER 65532:65532

# Expose metrics and health probe ports
EXPOSE 8080 8081

ENTRYPOINT ["/manager"]