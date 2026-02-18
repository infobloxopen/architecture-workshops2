FROM golang:1.25 AS builder

WORKDIR /app

# Cache module downloads
COPY go.mod go.sum* ./
RUN go mod download 2>/dev/null || true

# Copy source and build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /lab ./cmd/lab

# Final minimal image
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /lab /lab

ENTRYPOINT ["/lab"]
