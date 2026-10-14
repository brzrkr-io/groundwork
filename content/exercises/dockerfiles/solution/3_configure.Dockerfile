# Task 3: production defaults on top of the statusd:dev image from task 2.
#
#   docker build -f 3_configure.Dockerfile -t statusd:prod .
#   docker run --rm -p 8080:8080 statusd:prod

FROM statusd:dev

ENV APP_ENV=production PORT=8080
EXPOSE 8080
