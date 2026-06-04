FROM golang:1.25-alpine AS builder
WORKDIR /build
COPY backend/ .
RUN go mod download && \
    CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/server && \
    CGO_ENABLED=0 GOOS=linux go build -o /seed ./cmd/seed

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /server .
COPY --from=builder /seed .
COPY frontend/ ./frontend/
EXPOSE 8080
CMD ["./server"]
