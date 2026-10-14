# Task 1: the api image. CI builds it on every commit:
#
#   docker build -f 1_api.Dockerfile -t api:dev .
#   docker run --rm -p 8080:8080 api:dev

# TODO(1): Every build downloads all the Go modules again, even when a commit
# only changes one .go file. Make a code change reuse the cached download.
FROM golang:1.26
WORKDIR /src
COPY . .
RUN go mod download
RUN go build -o /usr/local/bin/api ./cmd/api
CMD ["api"]
