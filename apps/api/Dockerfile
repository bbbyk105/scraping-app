FROM golang:1.22-alpine

WORKDIR /app

# Install git and build dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build is done at runtime using go run for development
# For production, uncomment these lines:
# RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/server ./cmd/server/main.go
# CMD ["/app/bin/server"]

# For development, we use go run
CMD ["go", "run", "cmd/server/main.go"]
