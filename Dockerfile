# Build Stage
FROM golang:1.26-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o nmapshot main.go

# Package Stage
FROM alpine:latest

RUN apk add --no-cache nmap

WORKDIR /app

COPY --from=builder /app/nmapshot .

RUN chmod +x ./nmapshot

EXPOSE 8082

ENTRYPOINT ["./nmapshot"]
