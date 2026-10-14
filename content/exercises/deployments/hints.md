## Task 1: Keep three copies running

Three things are missing: how many copies, how the Deployment finds its
Pods, and the label each Pod gets. The last two must agree.

### Closer

`replicas:` and `selector:` sit directly under `spec:`. The selector is
`matchLabels:` with `app: web` indented under it. The Pod's labels go in
`template.metadata.labels`.

### Answer

```yaml
spec:
  replicas: 3
  selector:
    matchLabels:
      app: web
  template:
    metadata:
      labels:
        app: web
```

## Task 2: Fix the rejected Deployment

Read the error: which two things does it say don't match? The task says
which side to keep.

### Closer

Every `key: value` in `spec.selector.matchLabels` must also appear in
`spec.template.metadata.labels`. Change the template's `app` label.

### Answer

```yaml
  template:
    metadata:
      labels:
        app: api
```

## Task 3: Scale and relabel a running Deployment

Three edits in three places: the copy count, the Deployment's own labels,
and the Pods' labels. One place must not change at all.

### Closer

`replicas: 4`. `team: payments` goes under the top-level
`metadata.labels`. `track: stable` goes under `template.metadata.labels`,
next to `app: worker`. Leave `selector` alone: it is immutable, so changing
it makes the next apply fail with `field is immutable`.

### Answer

```yaml
metadata:
  name: worker
  labels:
    app: worker
    team: payments
spec:
  replicas: 4
  selector:
    matchLabels:
      app: worker
  template:
    metadata:
      labels:
        app: worker
        track: stable
```
