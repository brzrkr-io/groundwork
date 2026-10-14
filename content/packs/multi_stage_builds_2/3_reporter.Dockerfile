# Task 3: the reporter image.
#
#   docker build -f 3_reporter.Dockerfile -t reporter:dev .
#   docker run --rm reporter:dev

FROM golang:1.26 AS build
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/reporter ./cmd/reporter

FROM gcr.io/distroless/static-debian13
# TODO(3): The build fails with "pull access denied" for an image called
# builder. Once that's fixed, the container fails with
# exec: "/bin/sh": stat /bin/sh: no such file or directory.
COPY --from=builder /out/reporter /reporter
ENTRYPOINT /reporter
