# Task 1: the api service, compiled in a golang stage and shipped on distroless.
#
#   docker build -f 1_api.Dockerfile -t api:dev .
#   docker run --rm -p 8080:8080 api:dev

FROM golang:1.26 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/api .

FROM gcr.io/distroless/static-debian13
COPY --from=build /out/api /api
EXPOSE 8080
# TODO(1): Make the container run as a non-root user that Kubernetes can verify.
ENTRYPOINT ["/api"]
