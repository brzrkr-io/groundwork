# Task 2: the worker image. The worker binary is built by CI first, then:
#
#   docker build -f 2_worker.Dockerfile -t worker:dev .
#   docker run --rm worker:dev

FROM debian:trixie-slim

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates curl \
    && rm -rf /var/lib/apt/lists/*

COPY worker /usr/local/bin/worker
CMD ["worker"]
