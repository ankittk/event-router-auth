# event-consumer/Dockerfile

FROM golang:1.24.2 AS builder
WORKDIR /app

# Copy go.mod and go.sum first
COPY go.mod go.sum ./
RUN go mod download

# Copy all source code
COPY . .

WORKDIR /app/event-consumer
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /event-consumer main.go

# Final image — use static distroless
FROM gcr.io/distroless/static:nonroot
COPY --from=builder /event-consumer /event-consumer
USER nonroot:nonroot
ENTRYPOINT ["/event-consumer"]
