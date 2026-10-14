# Task 1: a throwaway image that prints the Go version of golang:1.26.
#
#   docker build -f 1_go_version.Dockerfile -t goversion .
#   docker run --rm goversion

FROM golang:1.26
CMD ["go", "version"]
