## Task 1: Give the web Pods an address

A Service has the same outline as every object: `apiVersion`, `kind`,
`metadata`, `spec`. In `spec` you need a selector and a list of ports. The
selector's labels come from the Pods, not from the Deployment.

### Closer

Services are in the core API, so `apiVersion: v1`. `selector:` is a plain
map under `spec` (no `matchLabels`). `ports:` is a list; one item with
`name`, `port` and `targetPort`.

### Answer

```yaml
apiVersion: v1
kind: Service
metadata:
  name: web
  namespace: shop
spec:
  selector:
    app: web
  ports:
    - name: http
      port: 80
      targetPort: http
```

## Task 2: Find the cart Pods

`Endpoints:` is empty, so the selector matches no Pod. Compare the
Service's `Selector:` line with the `LABELS` column of the Pods.

### Closer

The Pods have `app=cart,tier=backend`. Every label in the selector must be
on the Pods, with the same value.

### Answer

```yaml
  selector:
    app: cart
    tier: backend
```

## Task 3: Send search traffic to the right port

The endpoints end in `:8080`. The log says `"addr":":9898"`. The Service
sends traffic to its `targetPort`.

### Closer

Set `targetPort` to the container's port: the number 9898, or better its
name from `deploy/search.yaml`.

### Answer

```yaml
  ports:
    - name: http
      port: 80
      targetPort: http
```
