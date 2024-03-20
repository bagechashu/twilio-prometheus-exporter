# Use the official Golang image to create a binary
FROM golang:1.22-alpine AS builder

# Set the current working directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY ./src/go.mod ./src/go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source code into the container
COPY ./src/ .

ENV GOCACHE=/root/.cache/go-build

# RUN go mod tidy
# Build the Go app
RUN --mount=type=cache,target="/root/.cache/go-build" CGO_ENABLED=0 GOOS=linux  go build -o twilio_prometheus_exporter

# Start a new stage from scratch
FROM alpine:latest

# Set the current working directory inside the container
WORKDIR /root/

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/twilio_prometheus_exporter .

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
CMD ["./twilio_prometheus_exporter"]
