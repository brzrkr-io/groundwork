## Task 1: Make web Guaranteed

Requests and limits belong to the container, under one `resources` field.
Guaranteed needs both CPU and memory in both lists, with the same values.

### Closer

```yaml
          resources:
            requests:
              cpu: ...
              memory: ...
            limits:
              ...
```

### Answer

```yaml
          resources:
            requests:
              cpu: 250m
              memory: 128Mi
            limits:
              cpu: 250m
              memory: 128Mi
```

## Task 2: Stop the OOM kills

`OOMKilled` with exit code 137 means the kernel killed report for going
over its memory limit, 256Mi. The log says it holds 300,000,000 bytes:
about 286Mi, more than the limit.

### Closer

Raise `requests.memory` to 300Mi or more and `limits.memory` to 384Mi or
more, keeping the request at or below the limit. Leave the CPU request.

### Answer

```yaml
          resources:
            requests:
              cpu: 100m
              memory: 320Mi
            limits:
              memory: 400Mi
```

## Task 3: Get indexer scheduled

`Insufficient cpu` means no node has 1500m of unrequested CPU. Work out
the free CPU per node from `testdata/nodes.txt`: allocatable minus
requests.

### Closer

All three have 1930m allocatable. worker-1 has 550m free, worker-2 300m,
worker-3 800m. The request must be at most the largest of those.

### Answer

```yaml
            requests:
              cpu: 800m
              memory: 512Mi
```
