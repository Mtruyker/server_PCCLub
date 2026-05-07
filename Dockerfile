FROM golang:1.24-alpine AS build

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /pcclub-server ./cmd/server

FROM alpine:3.21

WORKDIR /app
COPY --from=build /pcclub-server /app/pcclub-server

ENV PORT=8080
ENV DATABASE_PATH=/app/data/pcclub.db

EXPOSE 8080
CMD ["/app/pcclub-server"]
