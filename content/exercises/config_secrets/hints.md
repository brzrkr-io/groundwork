## Task 1: Write the ConfigMap

Start from the four parts every object has. A ConfigMap keeps its settings
under `data`, not `spec`.

### Closer

`apiVersion: v1`, `kind: ConfigMap`, then `name` and `namespace` under
`metadata`. Under `data:`, one `KEY: value` per line. A value with spaces
needs no quotes, but quoting does no harm.

### Answer

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: web-config
  namespace: shop
data:
  PODINFO_LEVEL: debug
  PODINFO_UI_MESSAGE: hello from a ConfigMap
```

## Task 2: Fix the rejected Secret

`data` values must be base64. `YXBw` is; `correct-horse-battery` isn't
(base64 has no `-`). Either encode it, or put it where plain text is
allowed.

### Closer

Plain text goes under `stringData:`, a sibling of `data:`. Or run
`echo -n correct-horse-battery | base64` and paste the result under
`data`.

### Answer

```yaml
data:
  username: YXBw
stringData:
  password: correct-horse-battery
```

## Task 3: Wire the settings into the container

Two fields on the container, at the same level as `image`: one loads a
whole ConfigMap, the other is the usual `env` list with an entry whose
value comes from somewhere else.

### Closer

`envFrom:` is a list; its item has `configMapRef:` with a `name`. In
`env:`, the item for `DB_PASSWORD` has `valueFrom:` → `secretKeyRef:` with
`name` and `key`, instead of `value`.

### Answer

```yaml
          envFrom:
            - configMapRef:
                name: web-config
          env:
            - name: DB_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: db-credentials
                  key: password
```

## Task 4: Mount a config file

A mount needs two pieces: the Pod's `volumes` list says what the volume
is, and the container's `volumeMounts` says where it appears. The two are
joined by the volume's name.

### Closer

Under the Pod `spec` (next to `containers`): `volumes:` with an item
`name: config` and `configMap:` → `name: report-config`. Under the
container: `volumeMounts:` with `name: config`, `mountPath: /etc/podinfo`
and `readOnly: true`.

### Answer

```yaml
    spec:
      volumes:
        - name: config
          configMap:
            name: report-config
      containers:
        - name: report
          image: ghcr.io/stefanprodan/podinfo:6.9.2
          command: ["./podinfo", "--config-path=/etc/podinfo"]
          volumeMounts:
            - name: config
              mountPath: /etc/podinfo
              readOnly: true
```
