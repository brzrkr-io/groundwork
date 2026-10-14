# syntax=docker/dockerfile:1
# Task 3: the billing service. Some of its Go modules are private
# repositories under github.com/acme, so downloading them needs credentials.
#
#   docker build -f 3_private_deps.Dockerfile --secret id=netrc,src=$HOME/.netrc -t billing:dev .

FROM golang:1.26 AS build
ENV GOPRIVATE=github.com/acme/*
WORKDIR /src
COPY go.mod go.sum ./

# The credentials exist only while this one RUN step runs.
RUN --mount=type=secret,id=netrc,target=/root/.netrc go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /out/billing .

FROM gcr.io/distroless/static-debian13:nonroot
COPY --from=build /out/billing /billing
ENTRYPOINT ["/billing"]
