# Stage 1
FROM golang:1.24-alpine3.21 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main .






# Stage 2
FROM gcr.io/distroless/static:nonroot

WORKDIR /

COPY --from=builder /app/main .

USER nonroot

EXPOSE 8088 8089

ENTRYPOINT ["/main"]
