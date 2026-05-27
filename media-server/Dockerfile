FROM golang:1.21-alpine AS builder

RUN apk --no-cache add git

WORKDIR /app
COPY go.mod ./
RUN go mod download

COPY . .
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o media-server ./cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata ffmpeg
WORKDIR /app
COPY --from=builder /app/media-server .
COPY --from=builder /app/configs ./configs

EXPOSE 8081
CMD ["./media-server"]
