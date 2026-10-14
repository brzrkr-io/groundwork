# Task 1: the api image. CI builds it on every commit:
#
#   docker build -f 1_api.Dockerfile -t api:dev .
#   docker run --rm -p 8080:8080 api:dev

FROM golang:1.26
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /usr/local/bin/api ./cmd/api
CMD ["api"]
