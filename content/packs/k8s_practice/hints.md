## Task 1: Fix the Deployment

Two separate lines. Part 1's pull error says the tag doesn't exist; the
release list right below it says which does. Part 3's probe calls port
8080, and the log says where podinfo really listens.

### Closer

Change the image's tag after the `:`. In `readinessProbe.httpGet`, set
`port` to the container port's name (or its number), as the liveness probe
already does.

### Answer

```yaml
          image: ghcr.io/stefanprodan/podinfo:6.9.2
          # ...
          readinessProbe:
            httpGet:
              path: /readyz
              port: http
```

## Task 2: Fix the ConfigMap

The event says `couldn't find key level in ConfigMap notes/notes-config`.
The ConfigMap in part 2 has a key with another name.

### Closer

Rename the key in `data` to the one the Deployment's `configMapKeyRef`
asks for.

### Answer

```yaml
data:
  level: info
```

## Task 3: Fix the Service

First the selector: it must match the Pods' labels in part 4. Then the
target port: 80 is the port clients call, not the one podinfo listens on.

### Closer

`selector` → `app: notes`. `targetPort` → `http` (or 9898).

### Answer

```yaml
spec:
  selector:
    app: notes
  ports:
    - name: http
      port: 80
      targetPort: http
```
