# Stage 1: Build Go binary using official golang image
FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY . .

RUN go mod download

RUN go build -o main ./cmd

# Stage 2: Create small final image with only binary
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/main .

COPY .env .

EXPOSE 8080

CMD ["./main"]
