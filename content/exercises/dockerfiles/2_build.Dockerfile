# Task 2: build statusd from the source in this folder and run it.
#
#   docker build -f 2_build.Dockerfile -t statusd:dev .
#   docker run --rm -p 8080:8080 statusd:dev

# TODO(2): Start from golang:1.26, work in /src, copy the source in,
# build /usr/local/bin/statusd, and run statusd when the container starts.
