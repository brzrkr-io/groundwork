# Practice: deploy and debug a service

This morning's deploy of `notes` went badly: four problems, each hiding
behind the one before. You have the manifests that were applied and what
`kubectl` showed at each step. Find and fix all of them.

**New here:** nothing; practice reading `kubectl` output and fixing manifests.

## Learn

### Recap

| Symptom | Where to look | Exercise |
|---|---|---|
| `ImagePullBackOff` | describe: "Failed to pull image" | *Pod status: reading kubectl get* |
| `CreateContainerConfigError` | describe: Events | *ConfigMaps and Secrets: settings from the cluster* |
| `Running` but `0/1` | describe: "Readiness probe failed" | *Probes: readiness, liveness and startup* |
| Pods ready, Service unreachable | describe service: Selector, TargetPort, Endpoints | *Services: selectors, ports and endpoints* |

The triage order from *Debugging pods: describe, events and exit codes*:
`get` → `describe` (Events first) → `logs`.

### Why the old Pods are still there

In part 2 of the incident you'll see one new Pod failing next to the two
old ones. That's the rolling update from *Rollouts: surge, unavailable and
rollback*: with 2 replicas and the default 25%, `maxSurge` is 1 and
`maxUnavailable` is 0, so Kubernetes adds one new Pod and waits for it to
become ready before removing anything. A broken new version never replaces
the old one.

## Your task

Everything is in the namespace `notes`. `testdata/incident.txt` holds the
whole incident in four parts, in the order it happened. Read it first,
then fix each file. The tests check that the three manifests fit together.

### 1. Fix the Deployment

**Goal:** fix the two problems in `1_deployment.yaml`: the image and the
readiness probe.

- The image must be a tag that exists: the newest podinfo 6.9 tag listed
  in part 1.
- The readiness probe must reach the port podinfo listens on (part 3 has
  the log line that says which).

Once task 2's ConfigMap is fixed too, both Pods become ready:

```text
$ kubectl -n notes get pods
NAME                     READY   STATUS    RESTARTS   AGE
notes-59bb7b7554-bvdm8   1/1     Running   0          2m
notes-59bb7b7554-h2wxg   1/1     Running   0          2m
```

**Think it through:**
- What does "not found" in a pull error say about the tag?
- `connection refused` on 8080: which port is the app on, and what is it
  called in the container's `ports`?

### 2. Fix the ConfigMap

**Goal:** fix `2_configmap.yaml` so the Deployment finds the key it reads.
Keep the value `info`.

**Think it through:**
- Part 2's event names a key and a ConfigMap. Which key is in the
  ConfigMap, and which does the Deployment ask for?
- Which side is easier to change: one ConfigMap, or every Deployment that
  reads it?

### 3. Fix the Service

**Goal:** fix the two problems in `3_service.yaml` so
`curl http://notes.notes/` reaches the Pods. Keep `port: 80`.

**Think it through:**
- `Endpoints:` is empty. Compare the Selector with the Pods' labels.
- Once the endpoints appear, where would traffic go? Compare `TargetPort`
  with the port podinfo listens on.

## Run it for real

Optional, and worth it: replay the incident. Set up kind as in *Pods: your
first manifest*. From the exercise folder, with the **starter** files
first:

```bash
nix shell nixpkgs#kind
kind create cluster --name gw-k8s-practice
kubectl create namespace notes
kubectl apply -f 2_configmap.yaml -f 1_deployment.yaml -f 3_service.yaml
kubectl -n notes get pods -w                    # Ctrl+C when you've seen enough
```

Fix one file at a time, `kubectl apply -f` it, and watch the next problem
appear, as in the incident. When all three are fixed:

```bash
kubectl -n notes rollout status deployment/notes
kubectl -n notes describe service notes         # Endpoints: two IPs on :9898
kubectl -n notes run debug --rm -it --restart=Never --image=curlimages/curl:8.11.1 -- curl -sS http://notes/
kind delete cluster --name gw-k8s-practice
```

## Learn more

- [Kubernetes: Troubleshooting applications](https://kubernetes.io/docs/tasks/debug/debug-application/)
- [Kubernetes: Debug Services](https://kubernetes.io/docs/tasks/debug/debug-application/debug-service/)
