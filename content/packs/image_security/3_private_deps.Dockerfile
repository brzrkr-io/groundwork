# syntax=docker/dockerfile:1
# Task 3: the billing service. Some of its Go modules are private
# repositories under github.com/acme, so downloading them needs credentials.
#
#   docker build -f 3_private_deps.Dockerfile --build-arg GITHUB_TOKEN="$GITHUB_TOKEN" -t billing:dev .

FROM golang:1.26 AS build
ENV GOPRIVATE=github.com/acme/*
WORKDIR /src
COPY go.mod go.sum ./

# TODO(3): The GitHub token passed with --build-arg shows up in the build's
# history, its cache and the image's provenance. Download the private modules
# without a token in a build argument.
ARG GITHUB_TOKEN
RUN git config --global \
      url."https://x-access-token:${GITHUB_TOKEN}@github.com/".insteadOf "https://github.com/" \
 && go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /out/billing .

FROM gcr.io/distroless/static-debian13:nonroot
COPY --from=build /out/billing /billing
ENTRYPOINT ["/billing"]
