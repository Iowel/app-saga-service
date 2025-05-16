# Stage 1
FROM golang:1.24-alpine3.21 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o order_service ./main.go






# Stage 2
FROM gcr.io/distroless/static:nonroot

WORKDIR /app

COPY --from=builder /app/order_service .

EXPOSE 8086 8087

ENTRYPOINT ["/app/order_service"]
