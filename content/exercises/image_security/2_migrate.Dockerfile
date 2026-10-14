# Task 2: the migrate job. It applies the database migrations and exits.
#
#   docker build -f 2_migrate.Dockerfile -t migrate:dev .
#   docker run --rm -e DATABASE_PASSWORD migrate:dev

FROM golang:1.26 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/migrate .

FROM gcr.io/distroless/static-debian13:nonroot
COPY --from=build /out/migrate /migrate
ENV DATABASE_HOST=db.internal DATABASE_USER=migrate

# TODO(2): "docker history migrate:dev" shows the database password, and so
# does this file. Keep the password out of the image and out of the file.
# hunter2 is the staging password; production passes its own with --build-arg.
ARG DATABASE_PASSWORD=hunter2
ENV DATABASE_PASSWORD=${DATABASE_PASSWORD}

ENTRYPOINT ["/migrate"]
