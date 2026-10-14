# Task 2: build statusd from the source in this folder and run it.
#
#   docker build -f 2_build.Dockerfile -t statusd:dev .
#   docker run --rm -p 8080:8080 statusd:dev

FROM golang:1.26
WORKDIR /src
COPY . .
RUN go build -o /usr/local/bin/statusd .
CMD ["statusd"]
