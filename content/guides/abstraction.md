# Abstraction

**In one sentence:** abstraction is hiding the details of something behind a name and a promise, so you can use it without knowing how it works inside.

**Why it matters:** you type `kubectl apply -f deploy.yaml` and a Deployment appears; you never wrote the HTTP requests, the database writes or the scheduling decisions. Every tool you build is the same trick in miniature, and knowing how abstractions work (and how they break) is what lets you debug them when they fail.

**You'll learn:**

- What an abstraction is: a name, a promise and hidden details
- How a computer is built from layers, each one hiding the layer below
- The tools Go gives you to abstract, from constants up to interfaces and packages
- What a contract is: what callers may rely on, and what can change
- Why every abstraction leaks, and why the wrong one costs more than copy-paste
- Rules of thumb for when to abstract and when to wait

## You already use abstractions

### Driving a car

You press the accelerator and the car speeds up. You don't think about fuel injection, spark timing or gear ratios. You just trust the pedal.

The pedal is the part you touch: the *interface*. The engine is the part that does the work: the *implementation*. Between them is a promise: press harder, go faster.

Now drive an electric car. The implementation is completely different (no fuel, no gears), but the pedal and the promise are the same. You can drive it without retraining. That is the whole payoff of abstraction: **the inside can change without the outside noticing.**

### kubectl apply

Here is what one command hides:

```text
you type:     kubectl apply -f deploy.yaml
                  |
kubectl:      reads the YAML, finds the cluster in ~/.kube/config,
              sends an HTTPS request to the API server
                  |
API server:   checks who you are and what you may do, validates
              the object, saves it in etcd (the cluster database)
                  |
controllers:  notice the new Deployment, create a ReplicaSet,
              which creates Pods
                  |
scheduler:    picks a node for each Pod
                  |
kubelet:      on that node, asks the container runtime to pull
              the image and start the containers
```

You needed none of that to deploy. You needed one name (`apply`) and one promise ("the cluster will match this file").

### The three parts

| Part | Car | kubectl apply |
|---|---|---|
| A name | the accelerator | `kubectl apply` |
| A promise (the *contract*) | press harder, go faster | the cluster will match the file |
| Hidden details | engine, fuel, gears | HTTP, etcd, controllers, scheduler |

An *abstraction* is all three together. Another way to say it: abstraction is **choosing what you are allowed to ignore**. A good abstraction lets you ignore a lot, safely.

You have used abstractions in Go since your first program. `fmt.Println("hi")` hides how bytes reach your terminal. `strings.ToUpper("api")` hides how upper-casing works for every alphabet on Earth. You use them by name, trusting the promise.

## Layers: abstractions stacked on abstractions

Computers are built as a tower of abstractions. Each *layer* offers a simpler interface to the layer above, and hides the messy layer below.

```text
+-------------------+-----------------------------------------+
| the user          | types: deployctl rollout api            |
+-------------------+-----------------------------------------+
| your tool         | main: read the arguments, call functions|
+-------------------+-----------------------------------------+
| your functions    | rollout(), checkHealth(), notify()      |
+-------------------+-----------------------------------------+
| libraries         | fmt, strings, net/http                  |
+-------------------+-----------------------------------------+
| Go runtime        | memory, goroutines, garbage collection  |
+-------------------+-----------------------------------------+
| operating system  | files, processes, network sockets       |
+-------------------+-----------------------------------------+
| hardware          | CPU, RAM, disk, network card            |
+-------------------+-----------------------------------------+
```

Ideally, each layer talks only to the layer directly below it, through that layer's interface. Your function calls `fmt.Println`; it never talks to the terminal. `fmt` asks the operating system to "write these bytes to standard output". It doesn't know or care whether standard output is a terminal, a file, or a pipe into `grep`.

**Why build in layers?** Because you can swap a layer. The same Go source builds for Linux on an Intel server and macOS on an Apple laptop: the runtime and the operating system hide the difference. And because each layer keeps what you must think about small. While you write `rollout()`, the CPU is not your problem.

The *How a program runs* guide and *The operating system for programmers* guide look inside some of these layers.

## Go's abstraction tools, from small to big

Go gives you a ladder of tools. Each one hides more than the one before.

| Tool | What it hides | Example |
|---|---|---|
| Named constant | a magic value | `maxRetries` |
| Function | a sequence of steps | `podName(app, env, n)` |
| Struct type | how related data is bundled | `Host{Name, IP, Port}` |
| Method | behaviour that belongs to the data | `h.Addr()` |
| Interface | which type is doing the work | `Notifier` |
| Package | a whole set of internals | `net/http` |

### Named constants

```go
const maxRetries = 3
```

A bare `3` in the code says nothing. `maxRetries` says what the value *means*; the value itself becomes a detail you change in one place. It is the smallest abstraction there is.

### Functions

A function hides *how* something is done. The caller sees only the name, the parameters and the result.

```go
func podName(app, env string, n int) string {
	return fmt.Sprintf("%s-%s-%d", app, env, n)
}
```

Callers write `podName("api", "prod", 3)` and get `api-prod-3`. They don't know whether it uses `Sprintf` or `+`, and they don't need to. If the naming scheme changes, you change one function, not fifty call sites. The *Functions and decomposition* guide goes deeper.

### Struct types: data abstraction

Instead of passing four loose values (name, IP, port, zone) to every function, you define a `Host` struct and pass one thing. The struct is a name for "a thing made of these parts". This is *data abstraction*: you think in hosts, not in strings and numbers.

### Methods

A method attaches behaviour to the data it belongs to:

```go
func (h Host) Addr() string {
	return net.JoinHostPort(h.IP, strconv.Itoa(h.Port))
}
```

Callers ask a host for its address; they don't build it themselves. That matters more than it looks: an IPv6 address needs square brackets (`[2001:db8::1]:8080`), and simply gluing IP, colon and port gets that wrong. `net.JoinHostPort` adds the brackets when needed, and because only `Addr` builds addresses, that rule lives in exactly one place.

### Interfaces: behaviour abstraction

This is the most powerful tool in the list. A struct hides how *data* is laid out; an interface hides *which type* is doing the work. Callers depend only on what it can do. This is *behaviour abstraction*.

Your alerting code needs to tell someone that a host is down. Today that's Slack. Tomorrow it's email, and for the payments team, PagerDuty.

```go
package main

import "fmt"

// Notifier is anything that can deliver an alert message.
type Notifier interface {
	Notify(msg string)
}

type Slack struct{ Channel string }

func (s Slack) Notify(msg string) {
	fmt.Printf("[slack %s] %s\n", s.Channel, msg)
}

type Email struct{ To string }

func (e Email) Notify(msg string) {
	fmt.Printf("[email %s] %s\n", e.To, msg)
}

type PagerDuty struct{ Service string }

func (p PagerDuty) Notify(msg string) {
	fmt.Printf("[page %s] %s\n", p.Service, msg)
}

// alert knows only the Notifier promise, not who is behind it.
func alert(n Notifier, host string) {
	n.Notify(host + " is not responding")
}

func main() {
	notifiers := []Notifier{
		Slack{Channel: "#ops"},
		Email{To: "oncall@example.com"},
		PagerDuty{Service: "payments"},
	}
	for _, n := range notifiers {
		alert(n, "db-1")
	}
}
```

It prints:

```text
[slack #ops] db-1 is not responding
[email oncall@example.com] db-1 is not responding
[page payments] db-1 is not responding
```

Step by step:

- `Notifier` is the promise: "I have a `Notify(msg string)` method." It contains no code.
- `Slack`, `Email` and `PagerDuty` each have that method, so each one *is* a `Notifier`. Go needs no `implements` keyword: having the methods is enough.
- `alert` takes a `Notifier`. It calls `n.Notify(...)` without knowing which of the three it has.
- The loop passes each notifier to `alert` in turn, and each one runs its own `Notify`.

(Real notifiers would send HTTP requests and return an error; printing keeps the example small.)

Adding Microsoft Teams later means writing one new type with a `Notify` method. `alert` does not change at all. Neither do its tests. You practise this in *Interfaces* and *Interfaces: swap in a fake for tests*.

*Generics* are a cousin: one `Sum` function for both `int` and `float64` abstracts over the *type* of the data.

### Packages

A *package* is a folder of Go files that share a name, like `strings` or `net/http`. Names starting with a capital letter (`strings.ToUpper`) are *exported*, usable from outside; lowercase names are invisible outside the package. The exported names are the package's interface; everything else is implementation.

You'll build packages in the *Packages and project layout* module; the *Modularity and dependencies* guide is about how to split code into them.

## The contract: what callers may rely on

Every abstraction comes with a *contract*: the things callers may rely on. Everything else is an *implementation detail*: true today, but allowed to change.

Here is a function with its contract written in its doc comment:

```go
// retryDelay returns how many seconds to wait before retry number
// attempt (starting at 1). The delay grows with each attempt, up to
// a cap: it is never more than 60.
func retryDelay(attempt int) int {
	d := 1
	for i := 1; i < attempt && d < 60; i++ {
		d *= 2
	}
	if d > 60 {
		d = 60
	}
	return d
}
```

Callers may rely on "attempts count from 1, the delay grows, and it never goes above 60". The fact that it *doubles* is a detail. A caller that hard-codes "the third retry waits exactly 4 seconds" will break the day you add random jitter, and that's the caller's fault, not yours.

| You may rely on (contract) | Don't rely on (detail) |
|---|---|
| A Go map finds a value by its key | The order a map loops in |
| `kubectl get pods -o json` fields | The columns of the default table |
| A function's documented result | How it computes that result |
| An error you can check with code | The exact wording of the message |

The first row is a deliberate design choice: Go varies the order of map loops so nobody can depend on it. The second bites people every year: scripts that cut columns out of `kubectl get` break when a column is added, while the JSON stays stable.

Part of a contract is checked by the compiler: the function's name, parameter types and result types. The rest (what it actually does) lives in doc comments, and tests check it. See the *Testing and correctness* guide.

There's a hard truth here called Hyrum's law: with enough users, every observable behaviour of your system will be depended on by somebody, promised or not. So keep what's observable small.

## What abstraction buys you

**Change in one place.** Slack changes its webhook format. You edit `Slack.Notify`. Nothing else in the codebase moves.

**Test with fakes.** To test `alert`, you pass a fake notifier that just records the message in a slice. No real Slack channel gets spammed by your test suite, and the test runs in microseconds.

**Read less at once.** Compare reading 300 lines of SSH commands with reading this:

```go
for _, node := range nodes {
	drain(node)
	upgrade(node)
	verify(node)
}
```

You understand the plan in five seconds. Your working memory only holds a handful of things at once; abstraction is how you fit a big system into it.

## Every abstraction leaks

Joel Spolsky's *Law of Leaky Abstractions*: "All non-trivial abstractions, to some degree, are leaky." A *leaky abstraction* is one where the hidden details sometimes break through and you have to understand them after all.

- **TCP** promises a reliable stream of bytes over an unreliable network. Unplug a cable, and the promise can't be kept: you see timeouts, and you must think about the network after all.
- **ORMs** (libraries that let you treat database rows as objects) hide SQL. Then a loop over 1,000 objects quietly runs 1,000 queries, and you have to learn SQL to find out why the page takes 9 seconds.
- **kubectl** hides the whole cluster, until the API server is down and you get `The connection to the server ... was refused`, or a Pod sits in `Pending` because no node has enough free memory. Now you need to know about API servers, schedulers and resource requests.
- **Go strings** look like text, but they are bytes underneath:

```go
fmt.Println(len("cafe"), len("café")) // prints: 4 5
```

The `é` takes two bytes in UTF-8, and `len` counts bytes. The *Bytes, runes and UTF-8* exercises are about exactly this leak.

Spolsky's conclusion: abstractions save you time *working*, but they don't save you time *learning*. Good habit: **understand one layer below the one you use.**

### DevOps is built on leaky abstractions

| Abstraction | What it hides | How it leaks |
|---|---|---|
| Containers | the machine | share the host kernel; `OOMKilled` at the memory limit |
| Kubernetes | the servers | full nodes leave Pods `Pending`; a dead node takes local data |
| Terraform | cloud API calls | rate limits; drift from manual edits; surprise replacements |
| Cloud regions | data centres | a whole region can have an outage |

None of this makes these tools bad. It means that when one fails, the person who knows the layer below fixes it fastest.

## The costs of abstraction

### The wrong abstraction

You have two deploy scripts, one for `api` and one for `worker`. They look alike, so you merge them into one `deploy` function. Good.

Then `worker` needs database migrations, so you add a `migrate bool` parameter. Then `api` needs a canary, so you add `canary bool`. Six months later:

```go
func deploy(service string, migrate, canary, skipHealth,
	isWorker bool, region string)
```

The body is a forest of `if` statements. Every change for one service risks breaking the other.

Sandi Metz: "duplication is far cheaper than the wrong abstraction." Her advice when you find one: copy the shared code back into each caller, delete the parts each caller doesn't use, and look again. The right abstraction is often clear only once the duplication is back.

### Premature abstraction

Building a "pluggable multi-cloud layer" for a tool that only ever talks to AWS is *premature abstraction*: you pay today for a need that may never come, and guess its shape before you know it.

### Indirection

Each abstraction is a hop when reading. To find out what `alert` really does, you jump to `Notifier`, then hunt for the types that have a `Notify` method. An interface with one type behind it is all hop and no benefit. Abstraction adds *indirection*; make sure it pays for itself.

## Rules of thumb

- **Wait for the second or third repetition.** Copy once. When you write the same thing a third time, you know what's really shared.
- **Name things by what they mean, not how they work.** `inMaintenanceWindow(hour)` beats `hourBetween2And4(hour)`. The first survives a change of window; the second becomes a lie.
- **Keep interfaces small.** A Go proverb: "The bigger the interface, the weaker the abstraction." The standard library's `io.Writer` has one method, so files, network connections, memory buffers and HTTP responses can all be one. An interface with twelve methods has one implementation and is painful to fake.
- **Accept interfaces, return structs.** A function should take the smallest interface it needs (so callers can pass anything, including fakes), and return a concrete type (so callers get every field and method).
- **Hide what is likely to change.** Wrap the cloud API, the message format, the config file layout. Don't bother wrapping `fmt`.

## Abstraction, encapsulation, decomposition

Three ideas that are often mixed up. They work together, but each answers a different question.

| Idea | Question it answers | Example |
|---|---|---|
| *Decomposition* | How do I split this big job? | rollout = drain, upgrade, verify |
| Abstraction | What can I ignore when using a part? | `alert` ignores which notifier |
| *Encapsulation* | What is nobody allowed to touch? | an unexported `token` field |

You decompose a problem into parts, give each part an abstraction (a name and a contract), and encapsulate its details so nobody can come to depend on them.

## Common mistakes

- **Abstracting after one use.** You don't know what's shared yet. Wait for repetition.
- **Interfaces "just in case".** An interface with one implementation and no fake in the tests is pure indirection.
- **Relying on details.** Map order, the columns of `kubectl get`, the wording of an error message: none of them are promised.
- **Flag parameters piling up.** Every new `bool` parameter is a sign that one function is pretending to be several.
- **Treating an abstraction as magic.** When it leaks, you need the layer below. Learn it before the incident, not during.

## Check yourself

1. In `kubectl apply`, what is the interface, what is the promise, and name two hidden details.
2. A teammate's script reads the third column of `kubectl get pods` to find the pod status. Why is this fragile, and what should it use?
3. Why can `alert` work with a PagerDuty notifier that was written a year after `alert` itself?
4. Your `sendReport` function has grown parameters `toSlack bool, toEmail bool, compact bool, isWeekly bool`. What has gone wrong, and what could you do?
5. `len("naïve")` returns 6, not 5. Which abstraction leaked?
6. Which is better for a function that only needs to send messages: a parameter of type `Notifier` (one method) or `SlackClient` (a struct with 15 methods)? Why?

## Answers

1. The interface is the command `kubectl apply -f file`. The promise is that the cluster will be made to match the file. Hidden details include the HTTPS request to the API server, authentication, storing the object in etcd, controllers creating ReplicaSets and Pods, the scheduler choosing nodes and the kubelet starting containers.
2. The table layout is a detail meant for humans; a new column changes what "the third column" is. Use the contract instead: `-o json`, whose fields are part of the API.
3. Because `alert` depends only on the `Notifier` interface: "has a `Notify(msg string)` method". Any type with that method satisfies it, including types that didn't exist when `alert` was written. Go needs no `implements` declaration.
4. It's the wrong abstraction: one function pretending to be several. Copy it back into separate functions for each real use, delete what each doesn't need, and only re-extract what is truly shared.
5. The string abstraction: strings look like text, but `len` counts bytes, and `ï` takes two bytes in UTF-8.
6. `Notifier`. A small interface means any notifier, or a fake in tests, can be passed in. A big concrete type ties the function to Slack and makes it hard to test.

## Key terms

| Term | Meaning |
|---|---|
| abstraction | a name and a promise that let you use something while ignoring how it works |
| interface | the part of an abstraction you use; in Go, a type listing methods |
| implementation | the hidden code that actually does the work |
| contract | what callers may rely on; everything else may change |
| implementation detail | true today, not promised, may change |
| layer | one level in a stack of abstractions, using only the level below |
| data abstraction | grouping data under one name (a struct) so you think in things, not fields |
| behaviour abstraction | depending on what something can do (an interface), not what it is |
| leaky abstraction | one whose hidden details sometimes break through |
| premature abstraction | abstracting before you know what is really shared |
| package | a folder of Go files with one name; its exported names are its interface |
| exported | starts with a capital letter, so code outside the package can use it |
| indirection | an extra hop between the caller and the real work |
| decomposition | splitting a big job into smaller parts |
| encapsulation | stopping outside code from touching internal details |

## Go deeper

- [The Law of Leaky Abstractions](https://www.joelonsoftware.com/2002/11/11/the-law-of-leaky-abstractions/) (Joel Spolsky): the classic short essay, with TCP, SQL and network drives as examples.
- [The Wrong Abstraction](https://sandimetz.com/blog/2016/1/20/the-wrong-abstraction) (Sandi Metz): how a shared function decays into a flag-filled mess, and how to back out.
- [Go Proverbs](https://go-proverbs.github.io/) (Rob Pike): one-line design rules, including "the bigger the interface, the weaker the abstraction", each linked to the talk.
- [Effective Go: Interfaces](https://go.dev/doc/effective_go#interfaces): how Go's small, implicit interfaces are meant to be used.
- [Abstraction (computer science)](https://en.wikipedia.org/wiki/Abstraction_%28computer_science%29) on Wikipedia: levels of abstraction, data abstraction and the contract idea in more formal terms.
