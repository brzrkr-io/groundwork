# Task 3: the ingest image. You don't change this file: task 3 is its ignore
# file, 3_ingest.Dockerfile.dockerignore, which BuildKit finds by its name.
#
#   docker build -f 3_ingest.Dockerfile -t ghcr.io/acme/ingest:1.8.0 .

FROM golang:1.26 AS build
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build -o /out/ingest .

FROM gcr.io/distroless/static-debian13:nonroot
COPY --from=build /out/ingest /ingest
ENTRYPOINT ["/ingest"]
