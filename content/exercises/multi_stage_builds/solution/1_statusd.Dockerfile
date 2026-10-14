# Task 1: the statusd image, small enough to ship.
#
#   docker build -f 1_statusd.Dockerfile -t statusd:dev .
#   docker run --rm -p 8080:8080 statusd:dev

FROM golang:1.26 AS build
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build -o /out/statusd .

FROM gcr.io/distroless/static-debian13
COPY --from=build /out/statusd /statusd
ENTRYPOINT ["/statusd"]
