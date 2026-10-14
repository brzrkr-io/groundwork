# Task 3: the notifier image. notifier posts deploy messages to the team's
# chat webhook over HTTPS.
#
#   docker build -f 3_notifier.Dockerfile -t notifier:dev .
#   docker run --rm -e WEBHOOK_URL=https://chat.example.com/hooks/abc notifier:dev

# TODO(3): Write a build stage for a static notifier binary, and a final stage
# on scratch with the binary and the CA certificates.
