# Services: selectors, ports and endpoints

Pods come and go, and each new one gets a new IP address. A *Service* gives
a group of Pods one name and address that never changes. When "the app is
up but nothing can reach it", the Service is the first thing to check.

**New here:** Services, ClusterIP, `port`/`targetPort`, named ports, EndpointSlices, cluster DNS names, `kubectl port-forward`, debugging a Service with no endpoints.

## Learn

### A stable address in front of moving Pods

A Deployment's Pods are replaced on every rollout and every crash. Other
apps can't keep a list of their IPs. Instead they call a Service:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: docs
  namespace: wiki
spec:
  selector:            # send traffic to every Pod with these labels
    app: docs
  ports:
    - name: http
      port: 80         # the port clients call
      targetPort: web  # the container port the traffic goes to
```

The default type is *ClusterIP*: a virtual IP reachable only from inside
the cluster. (*NodePort* and *LoadBalancer* types also expose it outside;
you won't need them here.)

### Selectors and endpoints

A Service doesn't name its Pods. Like a Deployment, it has a *selector*,
and every **ready** Pod whose labels include all of the selector's labels
becomes an *endpoint*. Kubernetes keeps the list up to date in objects
called *EndpointSlices*:

```text
Service docs (10.96.12.40:80)  selector app=docs
   │
   ├──> 10.244.0.15:8080   Pod docs-5d9c7b-4kzrt   app=docs
   └──> 10.244.0.16:8080   Pod docs-5d9c7b-q8wtn   app=docs
```

If the selector matches nothing, the Service still exists and still has an
IP, but there is nowhere to send traffic. Connections fail straight away:

```text
curl: (7) Failed to connect to docs port 80 after 0 ms: Could not connect to server
```

Select on stable labels such as `app`, never on `pod-template-hash`: the
ReplicaSet adds that one and it changes on every rollout.

### Ports, by number or by name

| Field | Meaning |
|---|---|
| `port` | What clients connect to: `http://docs:80` |
| `targetPort` | Where the traffic goes in the Pod: a number, or the *name* of a container port |

A named `targetPort` (`web`, `http`) is looked up in each Pod's
`ports` list. If the app moves from 8080 to 9090, you change the container
port in one place and the Service follows.

A wrong `targetPort` is sneaky: the selector matches, so the endpoints
appear, but they point at a port nothing listens on, so connections are
refused.

### Names in cluster DNS

Every Service gets DNS names. From a Pod in the same namespace, `docs`
works. From anywhere in the cluster:

| Name | When to use it |
|---|---|
| `docs` | from the same namespace |
| `docs.wiki` | from another namespace |
| `docs.wiki.svc.cluster.local` | the full name |

### Debugging a Service

```bash
kubectl describe service docs     # Selector, TargetPort, Endpoints
kubectl get endpointslices -l kubernetes.io/service-name=docs
kubectl get pods -l app=docs --show-labels
kubectl logs deployment/docs      # which port does the app say it uses?
```

`Endpoints:` empty in `describe` means the selector matches no ready Pod:
compare the selector with the Pods' labels. Endpoints with the wrong port
means `targetPort` is wrong: compare it with the port in the app's logs.

To reach a Service from your own machine, forward a local port to it:

```bash
kubectl port-forward service/docs 8080:80   # then: curl localhost:8080
```

## Your task

Three Services in the namespace `shop`. Their Deployments are in `deploy/`;
read them, but don't change them. The tests check that each Service would
route to its Pods.

### 1. Give the web Pods an address

**Goal:** write `1_web_service.yaml`: a Service named `web`, in namespace
`shop`, for the Pods of `deploy/web.yaml`.

- Type ClusterIP (the default: you can leave `type` out).
- A selector that matches the web Pods' labels.
- One port: `name: http`, `port: 80`, and a `targetPort` that reaches the
  container port named `http` (9898).

```text
$ kubectl -n shop describe service web
Selector:     app=web
Port:         http  80/TCP
TargetPort:   http/TCP
Endpoints:    10.244.0.5:9898,10.244.0.6:9898
```

**Think it through:**
- Which labels do the Pods get? Look in `deploy/web.yaml`, under
  `template`.
- `targetPort` may be a number or a name. Which survives the app changing
  its port?

### 2. Find the cart Pods

**Goal:** fix the selector in `2_cart_service.yaml` so the Service routes
to the Pods of `deploy/cart.yaml`.

Nothing answers on `http://cart/`. `testdata/cart_debug.txt` has what
`kubectl` and `curl` showed. Leave the ports as they are.

**Think it through:**
- What does `Endpoints:` say in the `describe` output?
- Put the Service's selector next to the Pods' `LABELS` column. Which
  label doesn't match?

### 3. Send search traffic to the right port

**Goal:** fix `targetPort` in `3_search_service.yaml` so connections reach
podinfo.

This time `describe` shows endpoints, but `curl` still can't connect. The
evidence is in `testdata/search_debug.txt`.

**Think it through:**
- Which port do the endpoints point at? Which port does the app's log say
  it listens on?
- Could a name make this mistake impossible next time?

## Run it for real

Optional. Set up kind as in *Pods: your first manifest*, then from the
exercise folder:

```bash
nix shell nixpkgs#kind
kind create cluster --name gw-services
kubectl create namespace shop
kubectl apply -f deploy/
kubectl apply -f 1_web_service.yaml -f 2_cart_service.yaml -f 3_search_service.yaml
kubectl -n shop describe service web cart search    # every Endpoints: line has IPs
kubectl -n shop run debug --rm -it --restart=Never --image=curlimages/curl:8.11.1 -- curl -sS http://cart/
kubectl -n shop port-forward service/search 8080:80 # then curl localhost:8080 in a second terminal
```

Put the starter's bug back (`app: cart-api`), apply, and watch the
endpoints disappear in `describe`. Then clean up:

```bash
kind delete cluster --name gw-services
```

## Learn more

- [Kubernetes: Service](https://kubernetes.io/docs/concepts/services-networking/service/)
- [Kubernetes: DNS for Services and Pods](https://kubernetes.io/docs/concepts/services-networking/dns-pod-service/)
- [Kubernetes: Debug Services](https://kubernetes.io/docs/tasks/debug/debug-application/debug-service/)
