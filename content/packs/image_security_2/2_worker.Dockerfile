# Task 2: the worker image. CI builds it from github.com/acme/worker with
#
#   docker build -f 2_worker.Dockerfile \
#     --build-arg VERSION=1.8.0 --build-arg REVISION=$(git rev-parse HEAD) \
#     -t ghcr.io/acme/worker:1.8.0 .

FROM golang:1.26 AS build
ARG VERSION=dev
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags "-X main.version=${VERSION}" -o /out/worker .

FROM gcr.io/distroless/static-debian13:nonroot
COPY --from=build /out/worker /worker
# TODO(2): Finding 2: "for a worker image found running in a cluster, nobody
# could tell which repository, commit or version it was built from". Label it.
ENTRYPOINT ["/worker"]
