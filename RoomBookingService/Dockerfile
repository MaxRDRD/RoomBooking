FROM golang:1.26.1-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /app/bin/app ./cmd/app

FROM alpine:3.20

WORKDIR /app

COPY --from=builder /app/bin/app ./app

EXPOSE 8080

CMD ["./app"]
