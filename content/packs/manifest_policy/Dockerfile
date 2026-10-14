# kubepolicy in an image, for CI systems that run containers:
#   docker build -t kubepolicy:dev .
#   kubectl kustomize deploy/overlays/production | docker run --rm -i kubepolicy:dev -
FROM golang:1.26 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/kubepolicy ./cmd/kubepolicy

FROM gcr.io/distroless/static-debian13:nonroot
COPY --from=build /out/kubepolicy /kubepolicy
USER 65532:65532
ENTRYPOINT ["/kubepolicy"]
