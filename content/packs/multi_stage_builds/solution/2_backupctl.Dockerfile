# Task 2: the backupctl image. backupctl copies the team's databases to
# object storage every night.
#
#   docker build -f 2_backupctl.Dockerfile -t backupctl:dev .
#   docker run --rm backupctl:dev --help

FROM golang:1.26 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/backupctl ./cmd/backupctl

FROM gcr.io/distroless/static-debian13
COPY --from=build /out/backupctl /usr/local/bin/backupctl
ENTRYPOINT ["/usr/local/bin/backupctl"]
