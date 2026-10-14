# Task 2: the worker image. The worker binary is built by CI first, then:
#
#   docker build -f 2_worker.Dockerfile -t worker:dev .
#   docker run --rm worker:dev

FROM debian:trixie-slim

# TODO(2): Some builds fail with "E: Unable to locate package curl" or 404
# errors from the Debian mirror, and the image is 40 MB bigger than it needs
# to be. Make the package install reliable and small.
RUN apt-get update
RUN apt-get install -y ca-certificates curl

COPY worker /usr/local/bin/worker
CMD ["worker"]
