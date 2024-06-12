FROM golang:1.22

WORKDIR /app

COPY go.mod go.sum ./
RUN  go mod download
COPY client ./client
COPY test   ./test
