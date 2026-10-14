# Task 2: the backupctl image. backupctl copies the team's databases to
# object storage every night.
#
#   docker build -f 2_backupctl.Dockerfile -t backupctl:dev .
#   docker run --rm backupctl:dev --help

# TODO(2): Write the build stage: a static backupctl binary at /out/backupctl,
# built from ./cmd/backupctl, with the modules downloaded before the code is
# copied.

FROM gcr.io/distroless/static-debian13
COPY --from=build /out/backupctl /usr/local/bin/backupctl
ENTRYPOINT ["/usr/local/bin/backupctl"]
