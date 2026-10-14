# Task 1: the statusd image, small enough to ship.
#
#   docker build -f 1_statusd.Dockerfile -t statusd:dev .
#   docker run --rm -p 8080:8080 statusd:dev

FROM golang:1.26 AS build
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build -o /out/statusd .

# TODO(1): Add a runtime stage on distroless static that copies the binary
# out of the build stage and runs it when the container starts.
