FROM golang:1.21-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

# Build apenas da API (worker integrado)
RUN CGO_ENABLED=0 GOOS=linux GOGC=off GOMEMLIMIT=50MiB go build -ldflags="-w -s" -a -installsuffix cgo -o api ./cmd/api

FROM alpine:latest

RUN apk --no-cache add ca-certificates wget

RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

WORKDIR /app

COPY --from=builder /app/api .

COPY build/entrypoint.sh .

RUN chmod +x entrypoint.sh && \
    chown -R appuser:appgroup /app

USER appuser

EXPOSE 8080

CMD ["./entrypoint.sh"] 