# Task 3: production defaults on top of the statusd:dev image from task 2.
#
#   docker build -f 3_configure.Dockerfile -t statusd:prod .
#   docker run --rm -p 8080:8080 statusd:prod

FROM statusd:dev

# TODO(3): Set APP_ENV to production and PORT to 8080, and expose port 8080.
