# Task 3: the reportgen image. `go generate` runs stringer to write the
# String methods, then the build compiles reportgen:
#
#   docker build -f 3_tools.Dockerfile -t reportgen:dev .

FROM golang:1.26
RUN go install golang.org/x/tools/cmd/stringer@v0.52.0
WORKDIR /app
COPY go.mod go.sum /app/
RUN go mod download
COPY . /app/
RUN go generate ./...
RUN go build -o /usr/local/bin/reportgen ./cmd/reportgen
CMD ["reportgen"]
