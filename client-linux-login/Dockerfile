# Start from Go base image
FROM golang:1.24-alpine

# Set working directory
WORKDIR /app

# Copy Go code
COPY . .

# Build Go binary
RUN go build -o client

# Run binary when container starts
CMD ["./client"]
