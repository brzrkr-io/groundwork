# Task 3: the notifier image. notifier posts deploy messages to the team's
# chat webhook over HTTPS.
#
#   docker build -f 3_notifier.Dockerfile -t notifier:dev .
#   docker run --rm -e WEBHOOK_URL=https://chat.example.com/hooks/abc notifier:dev

FROM golang:1.26 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/notifier ./cmd/notifier

FROM scratch
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /out/notifier /notifier
ENTRYPOINT ["/notifier"]
