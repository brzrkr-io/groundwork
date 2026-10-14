## Task 1: Add a readiness probe

A probe is a field of the container, like `ports`. An HTTP probe says
which path to call and on which port.

### Closer

`readinessProbe:` at the same indentation as `ports:`, then `httpGet:`
under it, with `path:` and `port:` under that. The port can be the name
`http`.

### Answer

```yaml
          readinessProbe:
            httpGet:
              path: /readyz
              port: http
```

## Task 2: Stop the restarts

The events say `statuscode: 404`: the app is up and answering, but not on
that path. A healthy app being restarted is a probe problem, not an app
problem.

### Closer

The liveness probe calls `/health`. The last command in the fixture shows
which path podinfo really serves for liveness.

### Answer

```yaml
          livenessProbe:
            httpGet:
              path: /healthz
              port: http
```

## Task 3: Give search time to start

With the defaults, liveness gives up after `10 × 3` = 30 seconds of
`connection refused`, and the Pod is killed (exit code 137) before its
60-second start finishes. A startup probe holds liveness back until the
app is up.

### Closer

Add `startupProbe:` next to the other two, with the same `httpGet` as the
liveness probe, plus `periodSeconds` and `failureThreshold` whose product
is at least 120: for example 10 and 12.

### Answer

```yaml
          startupProbe:
            httpGet:
              path: /healthz
              port: http
            periodSeconds: 10
            failureThreshold: 12
```
