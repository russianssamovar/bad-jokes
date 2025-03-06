FROM golang:1.23.2 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod tidy

COPY . .

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

RUN go build -o badJokes .

FROM alpine:latest

WORKDIR /app

RUN apk --no-cache add ca-certificates

COPY --from=builder /app/badJokes .

COPY --from=builder /app/storage/migrations /app/storage/migrations

RUN chmod +x /app/badJokes

EXPOSE 9999

CMD ["/app/badJokes"]