FROM golang:1.21-alpine AS builder

RUN apk --no-cache add git

WORKDIR /app
COPY go.mod ./
RUN go mod download

COPY . .
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o api-server ./cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/api-server .
COPY --from=builder /app/configs ./configs

EXPOSE 8080
CMD ["./api-server"]
