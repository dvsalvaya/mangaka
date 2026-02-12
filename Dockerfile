# Build Stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
# -ldflags="-w -s" reduces binary size
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o mangaka cmd/main.go

# Final Stage
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies if any (e.g., certificates for https)
RUN apk add --no-cache ca-certificates

# Copy the binary from the builder stage
COPY --from=builder /app/mangaka .

# Create the downloads directory
RUN mkdir -p downloads

# Set the entrypoint
ENTRYPOINT ["./mangaka"]
