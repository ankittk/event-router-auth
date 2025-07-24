FROM golang:1.24.2 as builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o event-router ./cmd/server/main.go

WORKDIR /app/cmd/server
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /event-router main.go

# Final image (distroless static image)
FROM gcr.io/distroless/static:nonroot
COPY --from=builder /event-router /event-router
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/event-router"]
