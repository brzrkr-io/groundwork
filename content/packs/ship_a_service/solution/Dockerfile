# The image for inventoryd: compile it in a build stage, then run it from a
# small final stage as a non-root user.
#
#   docker build -t inventoryd:dev --build-arg VERSION=v1.4.0 .
#   docker run --rm -p 8080:8080 inventoryd:dev
#   docker run --rm inventoryd:dev -version

FROM golang:1.26 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG VERSION=dev
RUN CGO_ENABLED=0 go build -ldflags "-X main.version=${VERSION}" -o /out/inventoryd .

FROM gcr.io/distroless/static-debian13
COPY --from=build /out/inventoryd /inventoryd
USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/inventoryd"]
