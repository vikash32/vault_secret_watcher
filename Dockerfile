# Stage 1: Build the Go binary
FROM golang:latest AS build

# Set the working directory inside the container
WORKDIR /app

# Copy the Go application source code into the container

COPY . .

# Build the Go application with optimizations
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o vault-watcher .

# Stage 2: Create a minimal production-ready image
FROM alpine:latest

# Set the working directory inside the container
WORKDIR /app

# Copy the Go binary from the build stage
COPY --from=build /app/vault-watcher .

# Define the command to run your Go application
CMD ["./vault-watcher"]