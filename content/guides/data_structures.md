# Data structures

**In one sentence:** a data structure is a way of arranging data in memory so
that the operations you do most often are cheap.

**Why it matters:** "is `web-417` in our list of 20,000 hosts?" takes one
step with a map and up to 20,000 steps with a slice. Picking the right shape
for your data is often the whole difference between a script that finishes
instantly and one that times out in the pipeline.

**You'll learn:**

- Why every data structure makes some operations cheap and others expensive.
- How arrays, slices and maps are laid out, and what that means for speed.
- How to build a set, a stack and a queue from slices and maps.
- What linked lists, trees and graphs are, and where you meet them in ops.
- How to pick the right structure for a job.

## What a data structure is

Picture two ways to keep the on-call phone numbers for 500 services.

1. **A pile of sticky notes.** Adding one is instant: slap it on top. Finding
   the number for `payments` means flipping through notes until you hit it.
2. **A phone book sorted by name.** Adding one is slow: you have to squeeze it
   into the right place. Finding `payments` is fast: open near the middle,
   then jump left or right.

Same data, different arrangement, different costs. That is the whole idea.

A *data structure* is a way to arrange data so that certain operations are
cheap. No structure makes everything cheap, so you pick one by asking:
**what will I do most often with this data?**

| Operation | Example question |
|---|---|
| Get by position | "What is the 3rd step of the rollout?" |
| Find by key | "What port does `payments` listen on?" |
| Check membership | "Have I already pinged this host?" |
| Add or remove | "A new deploy job arrived; queue it." |
| Visit in order | "Print the hosts alphabetically." |

"Cheap" here means "the number of steps doesn't explode as the data grows".
The *Algorithms and Big-O* guide gives that idea a precise name. In this
guide, O(1) means "about the same number of steps no matter how big", and
O(n) means "steps grow in step with the number of items, n".

## Arrays: a fixed row of boxes

An *array* is a fixed number of boxes of one type: `[24]int` holds one counter
per hour of the day. You'll write them in *Arrays*, the first exercise of this
module.

In memory, an array is one *contiguous* block: the boxes sit next to each
other, with no gaps.

```text
  hours [24]int        each int is 8 bytes

  index:    0     1     2     3    ...   23
         +-----+-----+-----+-----+-----+-----+
         | 12  |  4  |  0  |  9  | ... | 31  |
         +-----+-----+-----+-----+-----+-----+
  byte:  0     8     16    24              184
```

**Why indexing is instant.** To find `hours[3]`, the computer does one sum:
start of the array + 3 × 8 bytes. It doesn't walk past boxes 0, 1 and 2. So
getting or setting any element by index is O(1), whether the array has 24
boxes or 24 million.

**The catch: the size is fixed.** The length is part of the type: `[24]int`
and `[25]int` are different types. An array can never grow. That's perfect
for "one slot per hour" or "one per weekday", and useless for "every host in
the inventory", which changes every week.

## Slices: an array that can grow

A Go *slice* is the answer to "I want array speed, but I don't know the size
yet". A slice is a small *header* with three fields that points at an array
hidden underneath, called the *backing array*.

```text
  jobs (slice header)
  +-----+-----+-----+
  | ptr | len | cap |
  |  o  |  3  |  4  |
  +--|--+-----+-----+
     |
     v   backing array
  +-------+-------+-------+-------+
  | build | test  | push  |       |   <- room for one more
  +-------+-------+-------+-------+
    0       1       2       3
```

- **len** is how many elements you're using.
- **cap** (capacity) is how many fit before the backing array is full.

Indexing works exactly like an array, so it's O(1).

### What `append` does when the array is full

When `len < cap`, `append` drops the new element into the next free box.
When `len == cap`, there is no room, so `append`:

1. allocates a new, bigger backing array (roughly double for small slices),
2. copies every existing element across,
3. adds the new element and returns a header pointing at the new array.

```go
package main

import "fmt"

func main() {
	var jobs []string
	for i := 1; i <= 9; i++ {
		jobs = append(jobs, fmt.Sprintf("deploy-%d", i))
		fmt.Println("len", len(jobs), "cap", cap(jobs))
	}
}
```

On a 64-bit machine with a current Go version, this prints:

```text
len 1 cap 2
len 2 cap 2
len 3 cap 4
len 4 cap 4
len 5 cap 8
len 6 cap 8
len 7 cap 8
len 8 cap 8
len 9 cap 16
```

Trace it: the capacity only changes at lengths 1, 3, 5 and 9, the moments the
array was full. The exact numbers are a runtime detail and can change between
Go versions; the pattern (grow by a multiple, not by one) is what matters.

**Why this is still fast.** Copying is slow, but because the capacity
doubles, copies get rarer as the slice grows. Spread over all the appends,
each `append` costs a small constant amount on average. This is called
*amortised* O(1).

**If you know the size, say so.** `make([]string, 0, 500)` creates an empty
slice with room for 500, so 500 appends never copy. You'll measure exactly
this in *Benchmarks: sub-benchmarks and preallocation*.

**Sharing warning.** Two slices can point at the same backing array, so a
change through one shows up in the other. *Slice gotchas: shared memory*
drills this, and the *Memory, pointers and sharing* guide explains why.

## Maps: find by name, not by position

A slice answers "what is at position 7?". Ops questions are usually "what is
the port for `payments`?". With a slice of names you'd check them one by one:
O(n). A *map* (in general CS, a *hash table*) answers it in O(1) on average.

### How a hash table finds a key

A *hash function* turns the key into a big number. That number picks a slot
(a *bucket*) in an internal array, and the value is stored there.

```text
  key "payments"  --hash-->  9182736455  --mod 8-->  slot 7
  key "auth"      --hash-->  2203948811  --mod 8-->  slot 3

  slots:  0    1    2    3        4    5    6    7
        +----+----+----+--------+----+----+----+------------+
        |    |    |    | auth:  |    |    |    | payments:  |
        |    |    |    | 8443   |    |    |    | 9000       |
        +----+----+----+--------+----+----+----+------------+
```

Looking up `payments` repeats the same sum, lands on slot 7, and checks the
key there. One hash plus a quick check, no matter how many keys there are.
When two keys land on the same slot (a *collision*), the map keeps both and
compares keys to tell them apart. Collisions are why it's "O(1) on average"
rather than "always exactly one step".

Go's real maps are cleverer than this drawing (since Go 1.24 they use a
design called Swiss tables), but the idea is the same.

### What maps cost you

- **No order.** Keys are scattered by their hash, so `range` over a map
  visits them in no useful order. Go even randomises it on purpose, so
  nobody relies on it. If you need order, sort the keys; *Sorting* shows how.
- **Keys must be comparable.** Strings, numbers and booleans work; slices
  can't be keys.
- **More memory** than a slice of the same items: the map keeps spare slots
  to stay fast.

### Ops example: count status codes

```go
package main

import (
	"fmt"
	"maps"
	"slices"
)

func main() {
	codes := []int{200, 200, 503, 404, 200, 503, 200}

	counts := map[int]int{}
	for _, c := range codes {
		counts[c]++
	}

	for _, c := range slices.Sorted(maps.Keys(counts)) {
		fmt.Println(c, counts[c])
	}
}
```

It prints:

```text
200 4
404 1
503 2
```

Step by step: `counts[c]++` reads the current count (0 for a missing key),
adds one and stores it. That's one O(1) map operation per log line, so the
whole count is O(n). Sorting the keys at the end makes the output stable.

## Sets: "have I seen this?"

A *set* is a collection where each item appears at most once, and the main
question is "is X in it?". Go has no set type; you use a map whose keys are
the items and whose values you barely care about.

### Ops example: dedupe hostnames

Inventory files from three teams list overlapping hosts. You want each host
once, in the order first seen.

```go
package main

import "fmt"

func main() {
	hosts := []string{"web-1", "db-1", "web-1", "cache-1", "db-1"}

	seen := map[string]bool{}
	var unique []string
	for _, h := range hosts {
		if seen[h] {
			continue
		}
		seen[h] = true
		unique = append(unique, h)
	}
	fmt.Println(unique)
}
```

It prints `[web-1 db-1 cache-1]`.

`seen[h]` is `false` for a host not yet in the map (the zero value), so the
first `web-1` is kept and the second is skipped. The map does the "have I seen
it?" check; the slice remembers the order.

You'll also see sets written as `map[string]struct{}`. `struct{}` is a value
that takes zero bytes, so it saves a little memory. Structs arrive in the
*Types and methods* module; `map[string]bool` is perfectly fine until then.

## Stacks and queues

Both are just "a list where you only touch the ends". The difference is
**which end you take from**.

```text
  STACK (last in, first out)        QUEUE (first in, first out)

   push -> | step 3 | -> pop        push -> | job 3 | job 2 | job 1 | -> pop
           | step 2 |                        back                 front
           | step 1 |
           +--------+
```

### Stack: undo in reverse

A *stack* is LIFO: last in, first out. Think of rolling back a failed
release: you applied database migrations 1, 2, 3, so you must undo 3, then 2,
then 1. Editor undo history and nested function calls work the same way.

With a slice, push is `append`, and pop takes the last element:

```go
package main

import "fmt"

func main() {
	var applied []string
	applied = append(applied, "add-users-table")
	applied = append(applied, "add-email-column")
	applied = append(applied, "add-index")

	for len(applied) > 0 {
		last := applied[len(applied)-1]
		applied = applied[:len(applied)-1]
		fmt.Println("rolling back", last)
	}
}
```

It prints:

```text
rolling back add-index
rolling back add-email-column
rolling back add-users-table
```

Both push and pop work at the end of the slice, so both are O(1).

### Queue: fair turns

A *queue* is FIFO: first in, first out. A deploy system that receives jobs
should run them in arrival order, like a line at a coffee shop.

```go
package main

import "fmt"

func main() {
	queue := []string{"deploy api", "deploy web"}
	queue = append(queue, "deploy worker") // a new job arrives at the back

	for len(queue) > 0 {
		job := queue[0]
		queue = queue[1:]
		fmt.Println("running:", job, "| waiting:", len(queue))
	}
}
```

It prints:

```text
running: deploy api | waiting: 2
running: deploy web | waiting: 1
running: deploy worker | waiting: 0
```

`queue[1:]` doesn't move anything: it makes a new slice header that starts
one box later, so taking from the front is O(1) too.

You'll meet queues everywhere later: channels in the *Concurrency* module
are queues between goroutines, and Kubernetes controllers pull work from a
workqueue in the *Kubernetes with client-go* module.

## Linked lists: know the idea, rarely use it

A *linked list* stores each item in its own little box (a *node*) together
with the address of the next node.

```text
  head
   |
   v
  +-------+---+    +-------+---+    +-------+-----+
  | web-1 | o----> | db-1  | o----> | web-2 | nil |
  +-------+---+    +-------+---+    +-------+-----+
```

**What it's good at:** inserting or removing a node in the middle is O(1)
once you're standing on it: change one arrow, move nothing.

**What it's bad at:** getting the 500th item means following 499 arrows,
O(n). The nodes are scattered around memory, so the CPU can't read ahead the
way it does along a slice's contiguous array.

In practice, Go programmers almost always use a slice instead. Modern CPUs
are so fast at walking contiguous memory that a slice usually wins even at
"insert in the middle". Go has a linked list in its standard library
(`container/list`), and you'll rarely see it in real code. It's still worth
knowing, because the idea of "a node pointing to the next node" is how trees
and graphs are built.

## Trees: things inside things

A *tree* is a structure where every item (*node*) has exactly one parent,
except the top one, the *root*. Nodes with no children are *leaves*. You use
trees every day.

**The filesystem.**

```text
  /                          <- root
  +-- etc
  |   +-- nginx
  |       +-- nginx.conf     <- leaf
  +-- var
      +-- log
          +-- syslog
          +-- nginx
              +-- access.log
```

**JSON and YAML documents.** A Kubernetes manifest is a tree: the root has
`metadata` and `spec`; `spec` has `template`; `template` has `containers`,
and so on. "The image of the first container" is a path from the root down.

**DNS.** Names are read right to left from the root: `.` then `com` then
`example.com` then `api.example.com`. Each level points to whoever answers
for the level below it: the `com` servers know who runs `example.com`.

Trees make "find everything under here" natural (delete a directory, read a
whole `spec` block). Sorted trees also keep data in order while staying fast
to search, which is how many databases index their tables.

## Graphs: things connected to things

Drop the "exactly one parent" rule and you get a *graph*: *nodes* connected
by *edges*. When edges have a direction ("A depends on B"), it's a *directed
graph*.

**Service dependencies.**

```text
  web ----> api ----> postgres
             |
             +------> redis <---- worker
```

Questions you'll ask about it:

- "If `redis` goes down, what breaks?" Follow the arrows backwards from
  `redis`: `api` and `worker`, then `web` through `api`.
- "In what order do I start things?" Dependencies first: `postgres` and
  `redis`, then `api` and `worker`, then `web`.
- "Is there a *cycle*?" If `postgres` somehow depended on `web`, nothing
  could start first. Build tools and Terraform refuse cycles for this reason.

**Kubernetes owner references.** A Deployment owns ReplicaSets, which own
Pods. Each child records its owners in `metadata.ownerReferences`. When you
delete the Deployment, Kubernetes follows those edges to clean up the
ReplicaSets and Pods. Usually this looks like a tree, but an object may list
more than one owner, so in general it's a graph.

In Go you'd usually store a graph as a map from each node to its neighbours,
like `map[string][]string{"web": {"api"}, "api": {"postgres", "redis"}}`.

## Which one should I use?

| You need to... | Use | In Go |
|---|---|---|
| Keep items in order, read by position | dynamic array | `[]T` slice |
| Exactly N slots, known up front | array | `[N]T` |
| Look up a value by a name or ID | hash map | `map[K]V` |
| Check "seen it?" / remove duplicates | set | `map[T]bool` |
| Undo the last thing first | stack | slice, use the end |
| Process in arrival order | queue | slice, take the front |
| Nested data: files, configs, DNS | tree | nested maps, slices, structs |
| Things that depend on things | graph | `map[string][]string` |

When unsure, start with a slice. Switch to a map the moment you catch
yourself searching a slice for something by name.

## Common mistakes

- **Searching a slice inside a loop.** Checking "is this host in the
  allowlist?" with a loop over the allowlist, for every host, is slow with
  big lists. Build a set from the allowlist once, then check the map.
- **Relying on map order.** Output that "looked sorted" in testing comes out
  shuffled in production. Sort the keys when order matters.
- **Writing to a nil map.** `var m map[string]int` then `m["x"] = 1` panics.
  Create maps with `make` or `{}` first.
- **Using an array when the size can change.** `[10]string` for "the hosts"
  breaks the day there are 11. Use a slice.
- **Forgetting `append` returns the slice.** `append(jobs, "x")` on its own
  line does nothing useful. Write `jobs = append(jobs, "x")`.
- **Popping from the wrong end.** A queue that takes from the back runs the
  newest job first, so old jobs can wait forever.

## Check yourself

1. Why is `hours[20]` exactly as fast as `hours[2]` for an array?
2. A slice has `len 4` and `cap 4`. What does `append` have to do to add a
   fifth element?
3. You loop over a map of service names and print them. A teammate says the
   output is in a different order every run. Is that a bug in Go?
4. Your script must skip hosts it has already rebooted. Which structure
   tracks that, and what's the Go type?
5. Rolling back three config changes: stack or queue? And processing
   incoming webhook events fairly?
6. Name one tree and one graph you deal with in ops work.

## Answers

1. An array is one contiguous block, so the address of any element is the
   start plus index × element size: one sum, whatever the index.
2. The backing array is full, so `append` allocates a bigger array (about
   double for a small slice), copies the 4 elements over, adds the fifth,
   and returns a slice header pointing to the new array.
3. No. Map iteration order is unspecified and Go randomises it on purpose.
   Sort the keys first if you need a stable order.
4. A set, written in Go as `map[string]bool` (or `map[string]struct{}`).
   Checking `rebooted[host]` is O(1) on average.
5. Rollback is a stack: undo the newest change first. Webhook events are a
   queue: first come, first served.
6. Trees: the filesystem, a YAML manifest, DNS names. Graphs: service
   dependencies, Kubernetes owner references, network routes between
   regions.

## Key terms

| Term | Meaning |
|---|---|
| data structure | A way to arrange data so some operations are cheap. |
| contiguous | Stored side by side in memory, with no gaps. |
| array | A fixed number of same-type boxes in one contiguous block. |
| slice | A header (pointer, len, cap) over a backing array that can grow. |
| backing array | The real array a slice's elements live in. |
| capacity | How many elements fit before `append` must reallocate. |
| amortised O(1) | Usually one step; occasionally slow; cheap on average. |
| hash table / map | Finds values by key using a hash function to pick a slot. |
| hash function | Turns a key into a number used to choose a slot. |
| collision | Two keys landing in the same slot. |
| set | A collection with no duplicates; asks "is X in it?". |
| stack | Last in, first out (LIFO). |
| queue | First in, first out (FIFO). |
| linked list | Nodes that each point to the next node. |
| tree | Nodes with exactly one parent each, except the root. |
| graph | Nodes joined by edges, with no parent rule; may have cycles. |
| cycle | A path of edges that leads back to where it started. |

## Watch

- [Data Structures: Crash Course Computer Science #14](https://www.youtube.com/watch?v=DuDz6B4cqVc) · CrashCourse · 10 min: arrays, strings, stacks, queues, linked lists, trees and graphs in one fast, well-animated overview.
- [Data Structures: Hash Tables](https://www.youtube.com/watch?v=shs0KM3wKv8) · HackerRank · 6 min: how a hash table turns a key into a slot, which is why a Go map finds `payments` without searching.
- [Arrays vs Linked Lists - Computerphile](https://www.youtube.com/watch?v=DyG9S9nAlUM) · Computerphile · 30 min: timing both on real hardware shows why the array usually wins, the point of *Linked lists: know the idea, rarely use it*.
- [GopherCon 2016: Inside the Map Implementation - Keith Randall](https://www.youtube.com/watch?v=Tl7mi9QmLns) · Gopher Academy · 27 min: the deeper talk, a Go team member on how Go maps really work: buckets, hashing and growth.

## Go deeper

- [Go Slices: usage and internals](https://go.dev/blog/slices-intro): the
  Go team's own explanation of slice headers, backing arrays and growth.
- [Go maps in action](https://go.dev/blog/maps): creating, iterating and
  sorting maps, with the "use a map as a set" pattern.
- [Faster Go maps with Swiss Tables](https://go.dev/blog/swisstable): how
  Go's real hash table works, once the drawing above feels easy.
- [Hash table (Wikipedia)](https://en.wikipedia.org/wiki/Hash_table):
  hashing, collisions and load factors in general CS terms.
- [Owners and Dependents](https://kubernetes.io/docs/concepts/overview/working-with-objects/owners-dependents/):
  the Kubernetes object graph built from `ownerReferences`.
