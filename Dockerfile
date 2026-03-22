# Stage 1: Build
FROM golang:1.25.8-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o api-home-pay ./cmd/api/main.go

# Stage 2: Run
FROM alpine:3.21 AS runner

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

COPY --from=builder /app/api-home-pay .

USER appuser

EXPOSE 8080

ENV PORT=8080
ENV GIN_MODE=release

CMD ["./api-home-pay"]
