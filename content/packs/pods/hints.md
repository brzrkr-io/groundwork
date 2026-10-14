## Task 1: Write your first Pod

Each blank is a value after a `key:`. The README's task lists all four, in
the order they appear in the file.

### Closer

`kind` is the type of object, with a capital letter. The Pod's name sits
under `metadata`; the container's name sits in the `containers` list. Copy
the image exactly, tag included.

### Answer

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: hello
spec:
  containers:
    - name: hello
      image: ghcr.io/stefanprodan/podinfo:6.9.2
```

## Task 2: Label it and name its port

Labels describe who the Pod is, so they go in `metadata`, next to `name`.
The port belongs to the container, next to its `image`.

### Closer

`labels:` is a map, indented under `metadata:`, with one `key: value` per
line. `ports:` is a list under the container: each item starts with `- `
and has `name` and `containerPort`.

### Answer

```yaml
metadata:
  name: web
  labels:
    app: web
    tier: frontend
spec:
  containers:
    - name: web
      image: ghcr.io/stefanprodan/podinfo:6.9.2
      ports:
        - name: http
          containerPort: 9898
```

## Task 3: Configure it with a command and env

Both fields belong to the container, at the same indentation as `image`.
Because the image has no `ENTRYPOINT`, whatever you put in `command` is the
whole command line, program included.

### Closer

`command` is a list of strings: `["program", "flag"]`. `env` is a list of
maps, each `- name: ...` with `value: ...` under it.

### Answer

```yaml
    - name: greeter
      image: ghcr.io/stefanprodan/podinfo:6.9.2
      command: ["./podinfo", "--level=debug"]
      env:
        - name: PODINFO_UI_MESSAGE
          value: hello from Kubernetes
```

`command: ["./podinfo"]` with `args: ["--level=debug"]` works too.
