# Task 4: the cron image, which runs the team's nightly cleanup jobs.
#
#   docker build -f 4_cron.Dockerfile -t cron:dev .
#   docker run --rm cron:dev

FROM golang:1.26 AS build
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/cron ./cmd/cron

FROM gcr.io/distroless/static-debian13
COPY --from=build /out/cron /app/cron
ENTRYPOINT ["/app/cron"]
