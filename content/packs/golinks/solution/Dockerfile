# The golinks image: compile a static binary in a build stage, then run it
# from a small distroless stage as a non-root user.
#
#   docker build -t golinks:dev --build-arg VERSION=v0.1.0 .
#   docker run --rm -p 8080:8080 -e GOLINKS_DB=/tmp/golinks.db golinks:dev

FROM golang:1.26 AS build
WORKDIR /src
# Modules first: this layer is rebuilt only when go.mod or go.sum change.
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG VERSION=dev
# modernc.org/sqlite is pure Go, so the binary needs no C libraries.
RUN CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -X main.version=${VERSION}" -o /out/golinks ./cmd/golinks

FROM gcr.io/distroless/static-debian13:nonroot
COPY --from=build /out/golinks /golinks
USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/golinks"]
