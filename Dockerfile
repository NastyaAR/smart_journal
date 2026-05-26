FROM golang:alpine AS builder
ENV GOTOOLCHAIN=auto
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/api ./cmd/api/main.go


FROM alpine:3.19
RUN apk --no-cache add ca-certificates curl
WORKDIR /root/

COPY --from=builder /app/api .

COPY .env .env
EXPOSE 3000
CMD ["./api"]