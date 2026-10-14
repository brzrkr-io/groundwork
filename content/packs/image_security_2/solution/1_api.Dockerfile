# Task 1: the api image. CI builds and pushes it with
#
#   docker build -f 1_api.Dockerfile -t ghcr.io/acme/api:1.8.0 .

# The tags are for people; the digests are what Docker pulls.
FROM golang:1.26@sha256:727841129a8d3bbd6b99dc9f63427b62bb5a98f6e6e856a5b1399bf19152c03e AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/api .

FROM gcr.io/distroless/static-debian13:nonroot@sha256:f0b02dd57038ad9b032fcb6ad271ebeaf4cc1804bd067250270908b22f4b0114
COPY --from=build /out/api /api
USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/api"]
