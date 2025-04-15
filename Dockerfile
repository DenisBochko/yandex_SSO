# Stage 1: Builder
FROM golang:1.23.1 AS builder

WORKDIR /app

COPY . .

RUN go mod tidy && \
    CGO_ENABLED=0 GOOS=linux go build -o /app/bin/sso ./cmd/sso/main.go

# Stage 2: Run 
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/bin/sso /app/bin/sso
COPY --from=builder /app/config /app/config
COPY --from=builder /app/db /app/db

EXPOSE 50051

CMD ["/app/bin/sso"]
# CMD ["bash"]
