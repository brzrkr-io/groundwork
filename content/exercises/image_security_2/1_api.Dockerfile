# Task 1: the api image. CI builds and pushes it with
#
#   docker build -f 1_api.Dockerfile -t ghcr.io/acme/api:1.8.0 .

# TODO(1): Finding 1: "two builds of the same commit used different
# compilers and base layers, and nobody can say which". Pin the base images.
FROM golang:latest AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/api .

FROM gcr.io/distroless/static-debian13:latest
COPY --from=build /out/api /api
USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/api"]
