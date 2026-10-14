# Task 1: the api image.
#
#   docker build -f 1_api.Dockerfile -t api:dev .
#   docker run --rm -p 8080:8080 api:dev

FROM golang:1.26 AS build
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/api ./cmd/api

# TODO(1): The image is 900 MB and the security scanner reports dozens of CVEs
# in packages the api never uses.
FROM golang:1.26
COPY --from=build /out/api /api
ENTRYPOINT ["/api"]
