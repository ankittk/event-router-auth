FROM golang:1.24.2 AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

WORKDIR /app/event-consumer
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /event-consumer main.go

FROM gcr.io/distroless/static:nonroot
COPY --from=builder /event-consumer /event-consumer
USER nonroot:nonroot
ENTRYPOINT ["/event-consumer"]
