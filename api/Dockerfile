# Use the official Go image as a parent image
# We're using the latest available version as of September 2024
FROM golang:1.23-alpine

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

# Set the working directory inside the container
WORKDIR /app

# Copy the Go module files
COPY api/go.mod api/go.sum ./

# Download the dependencies
RUN go mod download

# Copy the rest of the application code
COPY api/ .

# Build the Go application with CGO enabled
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o main .

# Expose the port the app runs on
EXPOSE 8080

# Run the binary
CMD ["./main"]