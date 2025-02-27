FROM golang:1.23.2 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod tidy

COPY . .

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64
RUN go build -o badJokes .

FROM alpine:latest

RUN apk --no-cache add sqlite

WORKDIR /app

COPY --from=builder /app/badJokes /app/badJokes
COPY config ./config
COPY storage/badJokes.db /app/storage/badJokes.db

RUN chmod +x /app/badJokes

EXPOSE 9999

CMD ["/app/badJokes"]
