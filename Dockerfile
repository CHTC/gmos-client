FROM golang:1.21

WORKDIR /app

COPY go.mod go.sum ./
RUN  go mod download
COPY client ./client
COPY test   ./test
