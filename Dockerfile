FROM golang:1.20 AS builder

WORKDIR /app

COPY ./proxy ./proxy

WORKDIR /app/proxy
RUN go mod download
RUN go mod tidy

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o ./proxy .

FROM alpine:latest
RUN apk update

WORKDIR /app

COPY --from=builder /app/proxy/proxy .

COPY hugo ./hugo


CMD ["/app/proxy"]