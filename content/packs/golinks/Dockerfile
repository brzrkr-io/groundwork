# The golinks image: compile a static binary in a build stage, then run it
# from a small distroless stage as a non-root user.
#
#   docker build -t golinks:dev --build-arg VERSION=v0.1.0 .
#   docker run --rm -p 8080:8080 -e GOLINKS_DB=/tmp/golinks.db golinks:dev

# TODO(8): Write a two-stage Dockerfile: a build stage that compiles
# ./cmd/golinks with the version stamped in, and a distroless final stage
# that runs it as a non-root user on port 8080.
