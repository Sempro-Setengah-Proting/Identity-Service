FROM golang:alpine AS builder

WORKDIR /app

# Install dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 go build -o main main.go

# Run stage
FROM alpine:latest

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -S app && \
    adduser -S app -G app

# Copy the binary from the builder stage
COPY --from=builder /app/main .

COPY --from=builder /app/migrations ./migrations

# Expose port
EXPOSE 8080

USER app

# Run the application
CMD ["./main"]
