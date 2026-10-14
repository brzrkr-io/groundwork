## Task 1: Check the Go version

You need two instructions: one that picks the starting image and one that
says what runs when the container starts. Look at the table in *The
instructions you need*.

### Closer

`FROM` takes `name:tag`. `CMD` in exec form is a JSON list with the program
first and each argument after it: `CMD ["program", "argument"]`.

### Answer

```dockerfile
FROM golang:1.26
CMD ["go", "version"]
```

## Task 2: Build statusd

Five instructions, one per numbered step, in the same order. The build
happens with `RUN`, the start-up command with `CMD`.

### Closer

`WORKDIR /src` moves into the folder, and `COPY . .` copies the build
context to "here". Then `RUN` the exact `go build` command from the task,
and finish with `CMD ["..."]`.

### Answer

```dockerfile
FROM golang:1.26
WORKDIR /src
COPY . .
RUN go build -o /usr/local/bin/statusd .
CMD ["statusd"]
```

## Task 3: Configure it for production

`FROM` is done. Add one instruction for the two variables and one for the
port, below it.

### Closer

`ENV KEY=value OTHER=value` sets several variables on one line. `EXPOSE`
takes the port number alone.

### Answer

```dockerfile
FROM statusd:dev

ENV APP_ENV=production PORT=8080
EXPOSE 8080
```
