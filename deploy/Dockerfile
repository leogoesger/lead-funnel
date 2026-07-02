# Build stage
FROM golang:1.26-trixie AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

RUN go install github.com/ardanlabs/kronk/cmd/kronk@latest

RUN kronk libs --install --local --os linux --arch arm64 --processor cpu

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o api ./cmd/api

# Expose port
EXPOSE 8080

# Run the application
CMD ["./api"]
