## Task 1: Shrink the 900 MB image

Only the final `FROM` is wrong. The build stage needs Go; the image you ship
doesn't.

### Closer

Replace the second `FROM golang:1.26` with the distroless static image.
Keep the `COPY` and `ENTRYPOINT` lines.

### Answer

```dockerfile
FROM gcr.io/distroless/static-debian13
COPY --from=build /out/api /api
ENTRYPOINT ["/api"]
```

## Task 2: Fix: no such file or directory

The binary was built with cgo on, so it asks for a C loader that distroless
static doesn't have. Turn cgo off for the build.

### Closer

Put `CGO_ENABLED=0` in front of `go build` on the same line.

### Answer

```dockerfile
RUN CGO_ENABLED=0 go build -o /out/worker ./cmd/worker
```

## Task 3: Fix the stage name and the entrypoint

Two fixes. Compare the name after `AS` with the name after `--from=`, then
look at the `ENTRYPOINT`: it has no brackets.

### Closer

Make the `--from=` name match the stage. Write the entrypoint as a JSON
list: `ENTRYPOINT ["..."]`.

### Answer

```dockerfile
FROM gcr.io/distroless/static-debian13
COPY --from=build /out/reporter /reporter
ENTRYPOINT ["/reporter"]
```

## Task 4: Run the file you copied

Read the last argument of the `COPY`: that's where the binary is. The
`ENTRYPOINT` runs a different path.

### Closer

Change the `ENTRYPOINT` path to the `COPY` destination (or the destination
to the `ENTRYPOINT` path).

### Answer

```dockerfile
COPY --from=build /out/cron /app/cron
ENTRYPOINT ["/app/cron"]
```
