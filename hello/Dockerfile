# Use the official Golang image to build the application
FROM golang:1.23 AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy the Go module files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the Go application
RUN CGO_ENABLED=0 go build -o hello hello.go

# Use a minimal image to run the application
FROM gcr.io/distroless/base-debian11

# Set the working directory inside the container
WORKDIR /

# Copy the built binary from the builder stage
COPY --from=builder /app/hello .

# Expose the port the application runs on
EXPOSE 8080

# Command to run the application
CMD ["./hello"]