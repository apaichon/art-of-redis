# Use the official Golang image
FROM golang:1.21.6 AS builder

# Set the Current Working Directory inside the container
WORKDIR /goclient

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go app
RUN go build -o goclient ./goclient/cmd/main.go

# Start a new stage from scratch
FROM alpine:latest

# Set the Current Working Directory inside the container
WORKDIR /root/

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /goclient/goclient .

# Expose port 8080 to the outside world
EXPOSE 5000

# Command to run the executable
CMD ["./goclient"]