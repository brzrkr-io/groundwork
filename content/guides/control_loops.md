# Declarative systems and control loops

**In one sentence:** in a declarative system you write down what you want, and a control loop keeps comparing that with what really exists and nudges reality until they match.

**Why it matters:** Kubernetes, Terraform, Argo CD and Flux all work this way, and the controllers you'll write with client-go in the next module must too; once you think in loops, "why did my pod come back after I deleted it?" stops being a mystery.

**You'll learn:**

- The difference between imperative ("do these steps") and declarative ("make it so").
- What desired state and actual state are, and how a control loop closes the gap.
- Why reconcile code must be idempotent and level-triggered.
- How Kubernetes is built from many small controllers, fed by watches, caches and work queues.
- How controllers write back safely: conflicts, owner references and finalizers.
- What drift and GitOps are, and how they use the same loop.

## Two ways to ask for something

You need three copies of the `web` service running version 1.8.2. Here are two ways to ask.

**Imperative:** list the steps.

```bash
#!/bin/sh
# deploy-web.sh
ssh node1 'docker run -d --name web web:1.8.2'
ssh node2 'docker run -d --name web web:1.8.2'
ssh node3 'docker run -d --name web web:1.8.2'
```

**Declarative:** describe the result.

```text
apiVersion: apps/v1
kind: Deployment
metadata:
  name: web
spec:
  replicas: 3
  selector:
    matchLabels: {app: web}
  template:
    metadata:
      labels: {app: web}
    spec:
      containers:
        - name: web
          image: web:1.8.2
```

*Imperative* code says how: do this, then this, then this. *Declarative* configuration says what: this is the state I want. Something else works out the steps.

Now ask "what happens if...":

| Situation | Bash script | Deployment |
|---|---|---|
| Run it twice | Fails: container `web` exists | Nothing changes |
| node2 dies tomorrow | Two copies until someone notices | A new pod starts elsewhere |
| Change to 5 replicas | Write new steps for nodes 4 and 5 | Edit `replicas: 5` |
| Interrupted after node1 | Half-deployed, nobody knows | The loop carries on next pass |

The script only knows the world at the moment it runs. The Deployment is a standing order that something keeps enforcing. That "something" is a control loop.

*Terraform*, a tool that creates cloud infrastructure (servers, networks, storage) from files you write, is declarative too: `.tf` files describe the infrastructure you want, and `terraform plan` works out the steps. The difference is that Terraform only acts when you run it. Kubernetes acts all the time.

## Desired state and actual state

Every declarative system juggles two pictures:

- *Desired state*: what you asked for. In Kubernetes, the `spec` of an object. In Terraform, your `.tf` files.
- *Actual state*: what really exists right now. Running pods, real EC2 instances. In Kubernetes, objects report it in their `status`.

The whole job of the system is to make actual match desired, and keep it that way. The gap between them is the work still to do.

## The control loop

A thermostat is the classic example. You set it to 21 °C (desired). It reads the room: 18 °C (actual). It's too cold, so it turns the heating on. A while later it reads 21 °C and turns the heating off. It never stops checking.

That's a *control loop*:

```text
            +------------------------------------+
            |                                    |
            v                                    |
      +-----------+      +-----------+      +-----------+
      |  observe  | ---> |   diff    | ---> |    act    |
      +-----------+      +-----------+      +-----------+
      read actual        compare with       create, update
      state              desired state      or delete
```

Notice what the thermostat does *not* do. It doesn't remember "I turned the heat on 3 minutes ago, so it must be warm now". Each time round, it looks at the room again. That habit is what makes control loops robust.

In Kubernetes, one pass of this loop is called *reconciling*, and the function that does it is the *reconcile function*.

## A toy reconciler

Here is the loop in plain Go, with maps standing in for a cluster. Each map says how many replicas each app has.

```go
package main

import (
	"fmt"
	"sort"
	"strings"
)

// reconcile makes actual match desired and reports what it changed.
// It never asks how things got this way; it only compares the two.
func reconcile(desired, actual map[string]int) []string {
	var did []string
	for name, want := range desired {
		have, exists := actual[name]
		if !exists {
			did = append(did,
				fmt.Sprintf("create %s (%d replicas)", name, want))
		} else if have != want {
			did = append(did,
				fmt.Sprintf("scale %s %d -> %d", name, have, want))
		}
		actual[name] = want
	}
	for name := range actual {
		if _, wanted := desired[name]; !wanted {
			did = append(did, "delete "+name)
			delete(actual, name)
		}
	}
	sort.Strings(did) // map order is random; sort so the output is stable
	return did
}

func main() {
	desired := map[string]int{"web": 3, "worker": 2}
	actual := map[string]int{"web": 1, "old-cron": 1}

	for round := 1; round <= 3; round++ {
		if round == 3 {
			actual["web"] = 5 // someone started 2 extra pods by hand: drift
		}
		did := reconcile(desired, actual)
		if len(did) == 0 {
			did = []string{"nothing to do"}
		}
		fmt.Printf("round %d: %s\n", round, strings.Join(did, "; "))
	}
}
```

It prints:

```text
round 1: create worker (2 replicas); delete old-cron; scale web 1 -> 3
round 2: nothing to do
round 3: scale web 5 -> 3
```

Step by step:

1. **Round 1.** `worker` is desired but missing, so it's created. `web` exists with 1 but should have 3, so it's scaled. `old-cron` exists but isn't desired, so it's deleted.
2. **Round 2.** Same function, same desired state. Everything already matches, so it does nothing. That's the important line.
3. **Round 3.** Someone started two extra `web` pods by hand. The loop doesn't know or care who. It sees 5, wants 3, and fixes it.

The second loop over `actual` is the same "diff two maps" idea from *Maps: diff a service catalog*. You'll write the real thing in *Reconcile loops: desired vs actual*.

## Reconcile must be idempotent

Round 2 above is the property everything rests on. Running reconcile again on an already-correct world must change nothing. That's *idempotency*: doing it twice has the same effect as doing it once. The *Distributed systems basics* guide explains why it matters for retries; here it matters even more, because a controller reconciles the same object again and again:

- after every change to it,
- after every error (it retries),
- after a restart of the controller,
- and on a timer, just in case (a *resync*; client-go replays every object from its local cache).

So write reconcile as "ensure", never as "do":

| Not idempotent | Idempotent |
|---|---|
| Create the Service | Create the Service if it doesn't exist |
| Add label `team=infra` | Set label `team=infra` |
| Append a sidecar container | Ensure exactly one sidecar is present |
| Increment the retry counter in status | Write the observed count into status |

A controller may also crash halfway through a reconcile. With "ensure" steps, the next run just finishes the job. You'll practise this in *Reconcile loops: idempotent updates*.

## Level-triggered, not edge-triggered

Two ways to react to change:

- *Edge-triggered*: act on each event. "Pod web-a was deleted, so create one pod."
- *Level-triggered*: act on the current state. "I want 3 pods. I see 2. Create one."

They sound the same until something goes wrong. Suppose the controller was restarting when the delete event happened, or the watch connection dropped and the event was lost, or two deletes arrived at once and were merged.

```text
  edge-triggered:   event lost  ->  never acts  ->  2 pods forever
  level-triggered:  next look   ->  sees 2/3    ->  creates 1 pod
```

An edge-triggered controller is only as good as its event stream. A level-triggered one treats events as a hint ("something about `web` changed, go look") and then re-reads the whole state. Missed or duplicated events don't matter, because the decision is based on what *is*, not on what *happened*.

Kubernetes controllers are level-triggered. That's why your reconcile function receives just a key, like `default/web` (namespace, then name), and not "what changed".

## Convergence, not instant change

When you `kubectl apply` a new image, nothing is updated at once. Each controller takes a step, the next one notices, and so on. The cluster moves towards the desired state over seconds or minutes. That process is *convergence*.

It's eventual consistency, seen from the inside. While converging:

- `kubectl get` can show old pods and new pods together.
- `status` can lag `spec`. Kubernetes tracks this with two counters: `metadata.generation` goes up each time the spec changes, and a controller copies it to `status.observedGeneration` once it has acted on that version. If they differ, the controller hasn't caught up yet.
- Some states never converge: an image that doesn't exist, a pod that asks for more memory than any node has. The loop keeps trying and reports why in `status`. Reading `status` and events is how you debug it.

So "done" in a declarative system means "status says it has converged", not "the apply command returned". A deploy pipeline that waits on `kubectl rollout status` is waiting for convergence.

## Drift

*Drift* is when actual state moves away from desired state without the desired state changing. Someone clicks around in the cloud provider's web console, opens a firewall rule by hand, or runs `kubectl edit` during an incident.

What happens next depends on the system:

- **A continuous controller fixes it.** Start two extra `web` pods by hand and the ReplicaSet controller deletes the extras, like round 3 of the toy. But Kubernetes only defends the `spec` it holds: `kubectl scale` or `kubectl edit` changes that spec, so Kubernetes obeys the new value. Undoing that needs a loop that compares with git (GitOps, below).
- **A run-on-demand tool reports it.** `terraform plan` refreshes what's really there, compares it with your files and shows the difference. Nothing changes until someone applies.

*Drift detection* is that comparison on its own: list the differences and alert, without fixing. Once *Cloud fundamentals* has introduced AWS, you'll build a tiny `terraform plan` for it in *Drift detection: a tiny terraform plan*. Part of doing it well is deciding what not to count, like tags AWS adds itself.

## Kubernetes is a set of controllers

Kubernetes isn't one big program that "runs your app". It's an API server that stores desired and actual state, plus many small controllers, each owning one kind of object. The *Kubernetes from the user's side* guide named the parts; here is how they cooperate.

```text
 kubectl apply: Deployment "web", replicas 3, image web:1.8.2
     |
     v
 API server  (stores every object; all controllers talk only to it)
     |
     +--> Deployment controller
     |      "no ReplicaSet for this template" -> creates ReplicaSet
     +--> ReplicaSet controller
     |      "want 3 pods, have 0"             -> creates 3 Pods
     +--> scheduler
     |      "these Pods have no node"         -> assigns each a node
     +--> kubelet on each node
            "a Pod for me isn't running"      -> starts the containers,
                                                 reports status
```

Each box runs its own loop over its own objects. None of them calls another directly. They coordinate only by reading and writing objects in the API server. Change the image and the Deployment controller makes a new ReplicaSet, scales it up and the old one down: a rolling update, built entirely from small reconciles.

This is also why deleting a pod "doesn't work": the ReplicaSet controller sees 2 of 3 and makes a new one. To get rid of pods, change the desired state: scale the Deployment or delete it.

## How a controller watches cheaply

A controller that asked the API server for every pod, every second, would flatten a big cluster. Instead client-go gives you three pieces. This section is a preview: you'll build each one in the *Kubernetes with client-go* module, so skim it now and come back then.

```text
 API server
    |  1. list everything once, then watch for changes
    v
 informer ----> local cache (an in-memory copy of the objects)
    |
    |  2. on add / update / delete: put the key "default/web" on the queue
    v
 work queue   (the same key five times = one item)
    |
    |  3. a worker takes a key
    v
 reconcile("default/web")
    reads from the cache, writes changes to the API server
    error? -> put the key back, with backoff
```

- A *watch* is a long-lived request that streams changes ("added", "modified", "deleted"). The server closes watches regularly, and one that has fallen too far behind gets an `ERROR` event ("410 Gone"), so code must list again: *Watching resources: closed watches and ERROR events*.
- An *informer* combines list and watch to keep a *cache* up to date, so reads don't touch the API server. The cache can be slightly behind the server, which level-triggered code tolerates. See *Informers: list, watch and cache*.
- A *work queue* holds keys, not events. It removes duplicates, never gives the same key to two workers at once, and re-queues failures with backoff. See *Workqueues: keys, dedup and the worker loop*.

Notice how each piece serves the level-triggered idea: events only decide *which* object to look at; reconcile decides *what* to do by reading current state.

## Writing back safely

Reconcile reads from the cache but writes to the API server. Three fields in every object's `metadata` keep those writes safe:

| Field | What it does | What your controller does |
|---|---|---|
| `resourceVersion` | Changes on every write to the object | Gets a Conflict error if it writes from a stale copy |
| `ownerReferences` | Says "this object belongs to that one" | Sets it on objects it creates, so they're deleted with their owner |
| `finalizers` | Names that must all be removed before a delete completes | Cleans up outside things, then removes its own name |

**Conflicts.** Your controller and a teammate both read `web` while its `resourceVersion` is 41. The teammate saves first, making it 42. Your update still says 41, so the API server rejects it with a *Conflict* error (HTTP 409) rather than silently overwrite their change. That's *optimistic concurrency*: no locks, just "refuse writes based on stale reads". The level-triggered fix is simple: return the error, let the work queue retry, and the next reconcile reads fresh state.

**Owners.** Each Pod has an *owner reference* to its ReplicaSet, and each ReplicaSet one to its Deployment. Delete the Deployment and the *garbage collector* (another controller) deletes everything that pointed at it. Set owner references on what your controller creates and you need no delete code; you'll do it in *Project: Kubernetes controller*.

**Finalizers.** Some cleanup is outside the cluster: a cloud load balancer, a DNS record. A *finalizer* is a name in `metadata.finalizers`. While any remain, a delete only marks the object (it gets a `deletionTimestamp`). The controller sees the mark, cleans up, and removes its finalizer; then the object goes. If that controller is broken or gone, the object sits in `Terminating` forever: a classic incident.

## GitOps: the same loop, pointed at git

Where does the desired state come from? Often someone runs `kubectl apply` from a laptop or a CI job. Then the real desired state is "whatever was applied last", and nobody can easily see it.

*GitOps* moves the desired state into a git repository and runs one more control loop:

```text
   git repo (manifests)  --desired-->  +-------------------+
                                       | Argo CD or Flux   |
   cluster (live objects) --actual-->  | diff, then apply  |
                                       +-------------------+
```

Argo CD and Flux run in the cluster, keep comparing what's in git with what's live, and apply the difference. What you get:

- **Every change is a commit**: reviewed in a pull request, with an author and a timestamp. That's your audit trail.
- **Rollback is `git revert`.**
- **Drift is visible or undone.** A hand edit with `kubectl` shows up as "OutOfSync", or is put back automatically, depending on settings.
- **CI doesn't need cluster credentials.** The agent inside the cluster pulls; nothing outside pushes. See the *Security basics for automation* guide for why that matters.

The catch is the same as any controller: a manual fix during an incident will be undone unless you also commit it.

## Common mistakes

- **Writing reconcile as a script of steps.** "Create the Service" fails the second time. Write "ensure the Service exists and looks like this".
- **Trusting the event.** Using an event's old and new values to decide what to do. Use the event to pick a key, then read current state.
- **Keeping state in the controller's memory.** It's lost on restart. Everything you need should be readable from the cluster (or the cloud).
- **Fighting the controller.** Deleting pods or editing a ReplicaSet by hand. Change the object at the top, usually the Deployment.
- **Treating "applied" as "done".** Wait for status to show convergence.
- **Hand-editing a GitOps-managed cluster.** The fix vanishes at the next sync. Commit it.

## Check yourself

1. Your bash deploy script fails halfway after starting the app on two of five hosts. Running it again errors on the first host. What property is the script missing, and how would a declarative tool avoid the problem?
2. A controller misses a "pod deleted" event during a restart. Why doesn't a Kubernetes controller end up one pod short forever?
3. You delete a pod that belongs to a Deployment and a new one appears within seconds. Which controller made it, and how do you really remove the pods?
4. A teammate's reconcile does `pod.Labels["restarts"] = old + 1` each time it runs. What goes wrong?
5. Someone changes a cloud storage bucket's access policy by hand in the provider's web console. What will `terraform plan` show, and what would a GitOps-style continuous controller do instead?
6. What do a work queue's "dedup" and "requeue with backoff" each give a controller?
7. Your controller's update to a Deployment fails with a Conflict error: "the object has been modified". What happened, and what should reconcile do?
8. After `kubectl delete`, a custom resource has shown `Terminating` for an hour. What is the usual cause?

## Answers

1. Idempotency. Rerunning should finish the job, not fail on work already done. A declarative tool compares the desired five hosts with the two it finds and starts only the missing three.
2. It's level-triggered. On its next reconcile (after a restart, the next related event or a periodic resync) it counts the pods that exist, sees fewer than desired and creates more. It doesn't need the lost event.
3. The ReplicaSet controller: it saw fewer pods than its `replicas` and created one. Scale the Deployment to 0, or delete the Deployment; its owner references let the garbage collector remove its ReplicaSets and pods too. Change the desired state, not the actual.
4. It isn't idempotent. Reconcile runs many times for one real change (resyncs, retries, other updates), so the number climbs on its own. Set values computed from current state instead of adding to them.
5. The plan shows a difference: your files want one policy, the real bucket has another, and applying would change it back. A continuous controller would notice on its own and undo the change, without anyone running a command.
6. Dedup means a burst of events for one object becomes one reconcile, not fifty, and no two workers handle the same key at once. Requeue with backoff means a failed reconcile is tried again later, without hammering the API server.
7. Someone else wrote the object after your controller read it, so your copy's `resourceVersion` is stale. The API server refuses rather than overwrite their change (optimistic concurrency). Return the error so the work queue retries; the next reconcile reads the current object and works out the change again.
8. A finalizer is still on it. Deletion waits until every finalizer is removed, and the controller meant to clean up and remove it is failing or not running. Fix that controller; removing the finalizer by hand skips its cleanup, such as deleting a cloud load balancer.

## Key terms

| Term | Meaning |
|---|---|
| Imperative | Describing the steps to take |
| Declarative | Describing the result you want |
| Desired state | What you asked for (`spec`, `.tf` files) |
| Actual state | What really exists right now (`status`, live resources) |
| Control loop | Observe, diff, act, repeat, forever |
| Reconcile | One pass of the loop for one object |
| Idempotent | Doing it twice has the same effect as once |
| Edge-triggered | Acting on each change event |
| Level-triggered | Acting on the current state, whatever the events |
| Resync | Reconciling every object again on a timer; client-go replays them from its cache |
| Convergence | Actual state moving step by step towards desired |
| Drift | Actual state changing without desired state changing |
| Drift detection | Reporting the differences without fixing them |
| Watch | A streaming request for changes to objects |
| Informer | List plus watch that keeps a local cache current |
| Work queue | A deduplicating queue of object keys to reconcile |
| resourceVersion | A version stamp on each Kubernetes object that changes on every write |
| Optimistic concurrency | No locks: a write based on a stale version is rejected with a conflict, and the writer re-reads and retries |
| Owner reference | Metadata saying an object belongs to another, so it's deleted with its owner |
| Garbage collector | The Kubernetes controller that deletes objects whose owner is gone |
| Finalizer | A name on an object that delays its deletion until a controller finishes cleanup and removes it |
| GitOps | Git as the desired state, with a controller applying it |

## Watch

- [What is GitOps, How GitOps works and Why it's so useful](https://www.youtube.com/watch?v=f5EpcWp0THw) · TechWorld with Nana · 12 min: a clear overview of desired state in git and an agent that pulls it into the cluster, matching the GitOps section.
- [Kubernetes Design Principles: Understand the Why - Saad Ali, Google](https://www.youtube.com/watch?v=ZuIQurh_kDk) · CNCF [Cloud Native Computing Foundation] · 38 min: the best talk on why Kubernetes is declarative, level-triggered and built from small controllers that watch the API server.
- [Writing Kube Controllers for Everyone - Maciej Szulik, Red Hat (Beginner Skill Level)](https://www.youtube.com/watch?v=AUNPLQVxvmw) · CNCF [Cloud Native Computing Foundation] · 37 min: a beginner tour of a real controller's informers, listers and work queue, the pieces from "How a controller watches cheaply".

## Go deeper

- [Controllers](https://kubernetes.io/docs/concepts/architecture/controller/) (kubernetes.io): the official explanation, starting from the same thermostat.
- [Writing Controllers](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-api-machinery/controllers.md) (Kubernetes community): the rules real controller authors follow, including level-driven design and work queues.
- [Declarative Management of Kubernetes Objects Using Configuration Files](https://kubernetes.io/docs/tasks/manage-kubernetes-objects/declarative-config/) (kubernetes.io): what `kubectl apply` does with your manifests.
- [Argo CD documentation](https://argo-cd.readthedocs.io/en/stable/): a GitOps controller that compares git with the live cluster and syncs the difference.
- [Control loop](https://en.wikipedia.org/wiki/Control_loop) (Wikipedia): the idea from engineering, long before Kubernetes.
