# Use the official Go image for building the application
FROM golang:1.21-alpine AS builder

# Install build dependencies for CGO and tzdata for timezone support
RUN apk add --no-cache gcc musl-dev tzdata

# Set the working directory
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application with CGO enabled
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o main ./cmd/server/main.go

# Use a minimal base image for the final container
FROM alpine:latest

# Install runtime dependencies for CGO and tzdata for timezone support
RUN apk add --no-cache libc6-compat tzdata

# Set the working directory
WORKDIR /root/

# Copy the built binary from the builder stage
COPY --from=builder /app/main .

# Set the timezone to UTC (optional, can be changed)
ENV TZ=UTC

# Expose the application port
EXPOSE 8080

# Command to run the application
CMD ["./main"]