# ConfigMaps and Secrets: settings from the cluster

In *Container config: settings from the environment* you made one image
run anywhere by reading settings from the environment. On Kubernetes those
settings live in two objects, *ConfigMaps* and *Secrets*, and the Pod
spec says how they reach the container.

**New here:** ConfigMaps, Secrets (`data` in base64, `stringData`), `env` with `valueFrom`, `envFrom`, volumes and `volumeMounts`, `kubectl rollout restart`.

## Learn

### ConfigMaps

A *ConfigMap* holds plain settings as string keys and string values:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: docs-settings
  namespace: wiki
data:
  LOG_FORMAT: json
  MAX_UPLOAD_MB: "20"        # quoted: every value must be a string
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: docs-nginx
  namespace: wiki
data:
  nginx.conf: |              # a value can be a whole file
    server {
      listen 8080;
    }
```

Unquoted `20` is a number in YAML, and the API server rejects it: values
in `data` must be strings. YAML also reads `true`, `no` and `on` as
booleans, so quote them too.
The `|` starts a multi-line string; every indented line after it belongs
to the value. A line with just `---` separates two objects in one file.

### Secrets

A *Secret* looks the same, but is meant for passwords, tokens and keys.
Under `data`, every value must be *base64*-encoded:

```bash
echo -n 'hunter2' | base64      # aHVudGVyMg==   (-n: no trailing newline)
echo 'aHVudGVyMg==' | base64 -d # hunter2
```

Write plain text under `data` and `kubectl apply` fails with `illegal
base64 data at input byte N`. Or use `stringData`: plain text that the API
server encodes for you when it stores the Secret.

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: smtp
type: Opaque
stringData:
  password: hunter2
```

**Base64 is not encryption.** Anyone who can read the Secret can decode
it. Secrets are safer than ConfigMaps only because access to them can be
limited separately, and they are kept out of logs and `describe` output.
Never commit real secret values to Git: a YAML file with `stringData` is
as readable as a text file.

### Getting settings into the container

| You want | Write in the container | Result |
|---|---|---|
| one key as one variable | `env` with `valueFrom.configMapKeyRef` or `valueFrom.secretKeyRef` | `SMTP_PASSWORD=hunter2` |
| every key as variables | `envFrom` with `configMapRef` or `secretRef` | `LOG_FORMAT=json`, `MAX_UPLOAD_MB=20` |
| every key as a file | a `volume` in the Pod, and a `volumeMount` in the container | `/etc/nginx/conf.d/nginx.conf` |

```yaml
    spec:
      volumes:                         # Pod level: what can be mounted
        - name: nginx-config
          configMap:
            name: docs-nginx
      containers:
        - name: docs
          image: nginx:1.27
          envFrom:
            - configMapRef:
                name: docs-settings    # LOG_FORMAT, MAX_UPLOAD_MB, ...
          env:
            - name: SMTP_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: smtp
                  key: password
          volumeMounts:                # container level: where it appears
            - name: nginx-config
              mountPath: /etc/nginx/conf.d   # each key becomes a file here
              readOnly: true
```

A volume is declared once for the Pod and mounted by name into each
container that needs it. A mount hides whatever the image had in that
folder, so mount into a folder of its own. `envFrom` turns *every* key
into a variable, even `nginx.conf`: that's why the file lives in a
ConfigMap of its own.

> **Why is it written like this?**
> - **Kubernetes requires** a volume's `name` to match the `volumeMounts`
>   entry, and `valueFrom` to name both the object and the key.
> - **Convention:** environment variable names in capitals, file keys
>   named like the file they become.
> - **Your choice:** env for small values, files for whole config files and
>   for secrets that tools expect on disk, such as TLS keys.

### When things change, or go missing

- **Environment variables are read once**, when the container starts.
  After editing a ConfigMap, run `kubectl rollout restart deployment/docs`
  to replace the Pods. Mounted files update by themselves after a minute
  or so.
- **A missing ConfigMap, Secret or key** stops the container before it
  starts. `kubectl get pods` shows `CreateContainerConfigError`, and
  `kubectl describe pod` says exactly what's missing:

```text
Warning  Failed  8s (x10 over 105s)  kubelet  spec.containers{docs}: Error: couldn't find key log-level in ConfigMap wiki/docs-settings
```

## Your task

podinfo reads settings from variables starting with `PODINFO_`, and from a
file `config.yaml` in the folder given by `--config-path`. Configure it
without touching the image. Everything is in the namespace `shop`.

### 1. Write the ConfigMap

**Goal:** write `1_configmap.yaml`: a ConfigMap `web-config` in namespace
`shop` with two keys.

- `PODINFO_LEVEL`: `debug`
- `PODINFO_UI_MESSAGE`: `hello from a ConfigMap`

```text
$ kubectl -n shop get configmap web-config
NAME         DATA   AGE
web-config   2      5s
```

**Think it through:**
- Which top-level key holds a ConfigMap's settings? (Not `spec`.)
- Does `hello from a ConfigMap` need quotes in YAML?

### 2. Fix the rejected Secret

**Goal:** make `2_secret.yaml` valid, so `password` holds
`correct-horse-battery` and `username` still holds `app`.

The error from `kubectl apply` is at the top of the file.

**Think it through:**
- `YXBw` is `app` in base64. What is wrong with the password line?
- You can encode the value yourself, or move it to a section that takes
  plain text. Both pass.

### 3. Wire the settings into the container

**Goal:** in `3_web.yaml`, give the `web` container every key of
`web-config` as a variable, and `DB_PASSWORD` from the key `password` of
the Secret `db-credentials`.

```bash
kubectl -n shop exec deployment/web -- sh -c 'echo $PODINFO_UI_MESSAGE'
# => hello from a ConfigMap
```

**Think it through:**
- Which field loads all keys at once, and which sets one variable?
- `DB_PASSWORD` must not hold the password itself. Where does `valueFrom`
  point?

### 4. Mount a config file

**Goal:** in `4_report.yaml`, mount the ConfigMap `report-config` (in
`report-config.yaml`, already written) at `/etc/podinfo`, read-only.

- A Pod volume named `config` from the ConfigMap `report-config`.
- A `volumeMounts` entry in the container for `config`, at
  `/etc/podinfo`, with `readOnly: true`.

```bash
kubectl -n shop exec deployment/report -- ls /etc/podinfo
# => config.yaml
```

**Think it through:**
- Which part goes in the Pod's `spec`, and which in the container?
- How does Kubernetes know which volume a mount means?

## Run it for real

Optional. Set up kind as in *Pods: your first manifest*, then from the
exercise folder:

```bash
nix shell nixpkgs#kind
kind create cluster --name gw-config-secrets
kubectl create namespace shop
kubectl apply -f 1_configmap.yaml -f 2_secret.yaml -f report-config.yaml
kubectl apply -f 3_web.yaml -f 4_report.yaml
kubectl -n shop rollout status deployment/web
kubectl -n shop exec deployment/web -- sh -c 'echo $PODINFO_UI_MESSAGE $DB_PASSWORD'
kubectl -n shop get secret db-credentials -o jsonpath='{.data.password}' | base64 -d
kubectl -n shop exec deployment/report -- cat /etc/podinfo/config.yaml
kubectl -n shop exec deployment/report -- wget -qO- localhost:9898   # "message": "hello from a mounted file"
```

Change `PODINFO_UI_MESSAGE` in `1_configmap.yaml`, apply it, and check the
variable again: unchanged until `kubectl -n shop rollout restart
deployment/web`. Then clean up:

```bash
kind delete cluster --name gw-config-secrets
```

## Learn more

- [Kubernetes: ConfigMaps](https://kubernetes.io/docs/concepts/configuration/configmap/)
- [Kubernetes: Secrets](https://kubernetes.io/docs/concepts/configuration/secret/)
- [Kubernetes: Good practices for Kubernetes Secrets](https://kubernetes.io/docs/concepts/security/secrets-good-practices/)
