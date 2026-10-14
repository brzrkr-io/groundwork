## Task 1: Run as a user Kubernetes can check

The image says `USER app`, a name. `runAsNonRoot` can only check a number,
so give Kubernetes the number: the UID of `app` in `testdata/passwd.txt`.

### Closer

The line `app:x:100:101::/home/app:/sbin/nologin` has the UID 100 and the
group 101. Add `runAsUser: 100` next to `runAsNonRoot: true` in the Pod's
`securityContext` (`runAsGroup: 101` is optional).

### Answer

```yaml
      securityContext:
        runAsNonRoot: true
        runAsUser: 100
        runAsGroup: 101
```

## Task 2: Pass the restricted standard

There are four rules in the warning. Two can go on the Pod (`runAsNonRoot`,
`seccompProfile`); two must go on the container (`allowPrivilegeEscalation`,
`capabilities`). Add the UID too, or `runAsNonRoot` stops the start as in
task 1.

### Closer

A `securityContext:` under the Pod `spec` (next to `containers`) and
another under the container (next to `image`). `seccompProfile:` has a
`type:` under it; `capabilities:` has a `drop:` list.

### Answer

```yaml
    spec:
      securityContext:
        runAsNonRoot: true
        runAsUser: 100
        runAsGroup: 101
        seccompProfile:
          type: RuntimeDefault
      containers:
        - name: api
          image: ghcr.io/stefanprodan/podinfo:6.9.2
          ports:
            - name: http
              containerPort: 9898
          securityContext:
            allowPrivilegeEscalation: false
            capabilities:
              drop: ["ALL"]
```

## Task 3: Give upload a writable /tmp

The log says `open /tmp/...: read-only file system`. Mount something
writable over `/tmp` and leave the rest of the image read-only.

### Closer

`volumes:` under the Pod `spec` with `- name: tmp` and `emptyDir: {}`;
`volumeMounts:` under the container with `- name: tmp` and
`mountPath: /tmp`.

### Answer

```yaml
    spec:
      volumes:
        - name: tmp
          emptyDir: {}
      containers:
        - name: upload
          # ...
          volumeMounts:
            - name: tmp
              mountPath: /tmp
```
