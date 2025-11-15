# syntax=docker/dockerfile:1
FROM golang:1.25 as builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server ./cmd/server

FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=builder /app/server /app/server
COPY configs/ configs/
ENV REDIS__ADDR=redis:6379
EXPOSE 8080
ENTRYPOINT ["/app/server"]
