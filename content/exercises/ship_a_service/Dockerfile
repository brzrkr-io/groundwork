# The image for inventoryd: compile it in a build stage, then run it from a
# small final stage as a non-root user.
#
#   docker build -t inventoryd:dev --build-arg VERSION=v1.4.0 .
#   docker run --rm -p 8080:8080 inventoryd:dev
#   docker run --rm inventoryd:dev -version

# TODO(2): Write a two-stage Dockerfile: a "build" stage that compiles
# inventoryd with the version stamped in, and a distroless final stage that
# runs it as a non-root user.
