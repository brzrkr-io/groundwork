# Task 2: the worker image. The worker looks up queue hosts by name (package
# net).
#
#   docker build -f 2_worker.Dockerfile -t worker:dev .
#   docker run --rm worker:dev

FROM golang:1.26 AS build
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/worker ./cmd/worker

FROM gcr.io/distroless/static-debian13
COPY --from=build /out/worker /worker
ENTRYPOINT ["/worker"]
