# Modularity and dependencies

**In one sentence:** modularity is splitting a program into parts with small, clear connections between them, so each part can be understood, changed and tested on its own.

**Why it matters:** in a 3,000-line `main.go`, fixing the Slack message means scrolling past AWS calls and YAML parsing, and any change can break something unrelated. Split well, the same tool lets you replace the Slack code without reading the AWS code. And every library you import is someone else's code running in your pipeline with your credentials.

**You'll learn:**

- Why we split code: less to hold in your head, clear ownership, reuse and testing
- Cohesion and coupling, the two forces that decide where to cut
- A package's public surface versus its internals, and why small is better
- Which way dependencies should point, and why Go forbids import cycles
- How Go versions third-party code (semver, `go.mod`, `go.sum`), and what a dependency really costs

## Why split code at all

Picture `deployctl`, a tool that reads a host inventory, rolls out a new version over SSH, checks health, and posts to Slack. Version one is a single file. It works. Then it grows:

- To fix a typo in the Slack message, you scroll through 3,000 lines.
- Two teammates edit the same file and their changes collide.
- You can't test the inventory parser without SSH access to real servers.

Splitting the code fixes each of these.

| Reason | What it means for `deployctl` |
|---|---|
| Less to think about | to change Slack messages, read only the Slack part |
| Ownership | the platform team owns `rollout`, the chat-ops team owns `slack` |
| Reuse | another tool can import `inventory` without the rest |
| Testing | test `inventory` with a string, no servers needed |

In computer science, a *module* is any part with a boundary around it: a function, a package, a whole service. Go also uses the word for something specific: a *Go module* is a versioned collection of packages with a `go.mod` file at its root. This guide uses both, and says which.

## Cohesion: things that change together live together

*Cohesion* is how strongly the contents of one part belong together.

A package `slack` that holds the client, the message formatting and the webhook retries is highly cohesive. Everything in it is about one thing, and when Slack changes its API, you change one package.

A package `utils` holding `formatBytes`, `parseYAML`, `sendSlack` and `md5sum` has low cohesion. It's a drawer of unrelated things, and every part of the program imports it, so every change to it touches everything.

A quick test: **can you describe the package in one sentence without the word "and"?** The Go team's advice on package names says the same: avoid meaningless names like `util`, `common` and `misc`, because they tell readers nothing and grow without limit.

## Coupling: how much parts know about each other

*Coupling* is how much one part depends on the details of another.

Say the report code reaches into the AWS code and reads its raw API pages, its pagination tokens and its retry counters. When the AWS code changes, the report breaks. That's tight coupling.

Instead, the AWS code can return a plain `[]Host`, and the report takes a `[]Host`. Now the AWS code can change however it likes, as long as it keeps producing hosts.

```text
tight:  report --reads internals of--> aws (pages, tokens, retries)

loose:  aws --returns []Host--> main --passes []Host--> report
```

Common sources of tight coupling:

- Reading or writing another package's internal fields
- Shared global variables that several parts change
- Depending on the exact text of another part's error messages or output
- Two services that read and write the same database tables

**The goal: high cohesion, low coupling.** The question to ask is: if I change this part, how many other parts must change too? The smaller the answer, the better the split.

## The public surface

Every module has a *public surface*: the names other code can use. Everything else is internal.

In Go, the rule is one letter: names that start with a capital letter are *exported* (visible outside the package), and lowercase names are *unexported* (visible only inside). Here's `deployctl` with an `inventory` package:

```text
deployctl/
├── go.mod                module example.com/deployctl
├── main.go               package main
└── inventory/
    └── inventory.go      package inventory
```

`inventory/inventory.go`:

```go
// Package inventory reads host inventory files.
package inventory

import (
	"fmt"
	"strings"
)

// Host is one line of the inventory: a name and a role.
type Host struct {
	Name string
	Role string
}

// Parse reads lines like "web-1 web". Blank lines are skipped.
func Parse(text string) ([]Host, error) {
	var hosts []Host
	for i, line := range strings.Split(text, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		h, err := parseLine(line)
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", i+1, err)
		}
		hosts = append(hosts, h)
	}
	return hosts, nil
}

// parseLine is unexported: no code outside this package can call it,
// so it can be renamed or rewritten without breaking anyone.
func parseLine(line string) (Host, error) {
	fields := strings.Fields(line)
	if len(fields) != 2 {
		return Host{}, fmt.Errorf("want \"name role\", got %q", line)
	}
	return Host{Name: fields[0], Role: fields[1]}, nil
}
```

`main.go`:

```go
package main

import (
	"fmt"

	"example.com/deployctl/inventory"
)

func main() {
	hosts, err := inventory.Parse("web-1 web\n\ndb-1 db\n")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	for _, h := range hosts {
		fmt.Println(h.Name, "is a", h.Role, "server")
	}

	_, err = inventory.Parse("web-1 web\ncache-1\n")
	fmt.Println("error:", err)
}
```

`go run .` prints:

```text
web-1 is a web server
db-1 is a db server
error: line 2: want "name role", got "cache-1"
```

Step by step:

- `main` imports the package by its full path: the module path from `go.mod` plus the folder name.
- It can use `inventory.Parse` and `inventory.Host` because they start with capitals.
- The first call parses two hosts and skips the blank line.
- The second call hits `cache-1`, which has no role, and the error says which line.
- If `main` tried to call `inventory.parseLine`, the build would fail with `undefined: inventory.parseLine`.

**Every exported name is a promise.** Once other code uses it, changing it breaks them (see the *Abstraction* guide on contracts). So start small: you can always export something later, but un-exporting it is a breaking change.

Go adds one more wall: a package inside a folder named `internal` can only be imported by code in the tree rooted at the folder that contains `internal`. Use it for code that is shared inside your project but must not become anyone else's dependency. You practise both in *Exported and unexported names* and *internal packages*.

## Which way should dependencies point?

`deployctl` has an important part and a detailed part:

- **The policy:** "upgrade hosts one at a time; report progress; stop if a health check fails". This rarely changes, and it's the reason the tool exists.
- **The details:** Slack, SSH, AWS. These change all the time: Slack changes its API, your company moves to Teams.

If `rollout` imports `slack` directly, the important part depends on the detail. You can't test a rollout without Slack, and switching to Teams means editing the rollout logic.

Flip it. `rollout` defines the small interface it needs, and anything that satisfies it can be plugged in:

```go
// Package rollout upgrades hosts one at a time.
package rollout

// Notifier is all rollout needs from the outside world.
// Rollout owns this interface; it has never heard of Slack.
type Notifier interface {
	Notify(msg string)
}

// Run upgrades each host in order and reports progress.
func Run(hosts []string, n Notifier) {
	for _, h := range hosts {
		// ... drain, upgrade and verify h here ...
		n.Notify("upgraded " + h)
	}
	n.Notify("rollout finished")
}
```

A separate `slack` package has a `Client` type with a `Notify(msg string)` method, and `main` connects the two:

```go
rollout.Run(hosts, slack.Client{Channel: "#deploys"})
```

```text
before:  main --> rollout --> slack --> net/http
                  (rollout can't run, or be tested, without Slack)

after:   main --> rollout      (defines the Notifier interface)
           |
           +----> slack        (Client has a Notify method)

         main is the only place that knows about both
```

Now `rollout` imports nothing about chat. Its tests pass a fake notifier that records messages in a slice. Teams support is a new package; `rollout` doesn't change.

The rule: **important, stable code should not import detailed, changeable code.** Put an interface at the boundary, owned by the code that uses it. This is called *dependency inversion*, and you met the testing side of it in *Fakes and dependency injection*.

Kubernetes is built this way. The kubelet doesn't import containerd or CRI-O. It talks to any container runtime through the *Container Runtime Interface* (CRI), so runtimes can be swapped without changing Kubernetes. Storage (CSI) and networking (CNI) plugins work the same way.

## Dependency graphs and cycles

Draw an arrow from each package to every package it imports, and you get a *dependency graph*:

```text
            main
          /  |   \
         v   v    v
   rollout  slack  inventory
         \           ^
          +----------+
     (rollout also imports inventory)
```

`go list -deps .` prints every package your program depends on, all the way down.

Now suppose `rollout` imports `notify` to send messages, and `notify` imports `rollout` to read a `rollout.Status`. That's a *cycle*. Go refuses to build it:

```text
package example.com/cyc
	imports example.com/cyc/rollout from main.go
	imports example.com/cyc/notify from rollout.go
	imports example.com/cyc/rollout from notify.go: import cycle not allowed
```

### Why Go forbids import cycles

- **Build order.** Go compiles each package after the packages it imports, and can reuse the compiled result of any package that hasn't changed. In a cycle, there is no package to compile first: the whole loop would have to be built as one unit.
- **Start-up order.** Package-level variables and `init` functions (setup code that runs before `main`) are set up in dependency order: a package's imports go first. A cycle makes that order ambiguous.
- **Design.** Two packages that need each other are really one package in two folders. You can't understand, test, reuse or release either one alone. Go's designers say cycles, where allowed, end up tangling large parts of a codebase together.

### Breaking a cycle

- **Move the shared piece down.** Put `Status` in a small third package that both import.
- **Use an interface.** Let `rollout` define a `Notifier` interface, as above, so it no longer imports `notify`.
- **Merge them.** If they really are one thing, make them one package.

## Third-party dependencies

So far the dependencies were your own packages. Third-party ones are code written by strangers.

### Semantic versioning

Go modules use *semantic versioning*: a version like `v1.4.2` is `MAJOR.MINOR.PATCH`.

| Bump | Example | Promise |
|---|---|---|
| Patch | `v1.4.2` to `v1.4.3` | bug fixes only |
| Minor | `v1.4.3` to `v1.5.0` | new features, nothing removed |
| Major | `v1.5.0` to `v2.0.0` | may break your code |
| `v0.x.y` | `v0.37.1` | no promises; anything may change |

In Go, a new major version also changes the import path: `example.com/lib/v2` is a different path from `example.com/lib`, so both can exist in one build.

Semantic versioning is a promise made by humans. Sometimes a "patch" release breaks things anyway. That's why you test upgrades instead of trusting the number.

### go.mod and go.sum

`go.mod` names your module and lists the minimum version you need of each dependency:

```text
module example.com/deployctl

go 1.26

require (
	k8s.io/client-go v0.37.1
	sigs.k8s.io/yaml v1.6.0
)
```

(Real files also list `// indirect` lines: the dependencies of your dependencies.) Notice `client-go` is still `v0.x`: Kubernetes' Go client makes no semver promise, and its APIs do change between minor versions.

When two packages need different versions of the same module, Go picks the *highest of the versions anyone asked for*, never something newer. The build is the same today and next month, even if new versions come out. This is called *minimal version selection*.

`go.sum` holds a cryptographic hash of each module version you use (plus a second line per version, for its `go.mod` file):

```text
sigs.k8s.io/yaml v1.6.0 h1:G8fkbMSAFqgEFgh4b1wmtzDnioxFCUgTZhlbj5P9QYs=
```

If a download ever doesn't match its hash, the build fails. Go also checks new modules against a public checksum database, so everyone in the world gets the same bytes for the same version. Commit both files.

```bash
go get sigs.k8s.io/yaml@v1.6.0   # add or change a dependency
go mod tidy                      # add what's missing, drop what's unused
```

### Supply-chain risk

Every dependency is code you run with your tool's permissions. In CI that often means deploy credentials and cloud keys. A *supply-chain attack* is getting malicious code into your program through something you depend on: a hijacked maintainer account publishing a bad version, a look-alike package name, or a patient attacker who becomes a maintainer. In 2024 a backdoor was found in xz, a compression library shipped in many Linux distributions, after one such long campaign.

Go makes some attacks harder: `go.sum` and the checksum database catch altered downloads, versions are fixed so builds don't silently upgrade, and fetching or building a module never runs its code. Tools like `govulncheck` report known vulnerabilities in the dependencies your code actually calls. None of this protects you from a dependency that is malicious from the start.

### What a dependency costs

A Go proverb: "A little copying is better than a little dependency."

Importing a library for one ten-line function brings the whole library, and *its* dependencies, into your build. You now inherit their bugs, their security issues and their upgrade schedule. In 2016, an 11-line npm package called `left-pad` was unpublished, and builds across the JavaScript world broke within hours.

Before adding a dependency, ask:

- How much of it will I use? Could I write that part in an afternoon?
- Does the standard library already do it? Go's covers HTTP, JSON, flags, templates, crypto and testing.
- Is it maintained? When was the last release?
- How many dependencies does it pull in? (`go mod graph` shows them.)
- Is its license one my company can use?

The opposite mistake is real too. Don't write your own cryptography, YAML parser or Kubernetes client. When the problem is hard and the library is well-maintained, depend on it.

## The same trade-offs, one level up

Microservices and shared libraries are modules at a bigger scale, with the same rules:

| In code | In a platform |
|---|---|
| a package | a service |
| exported names | the service's API |
| a major version for breaking changes | a `/v2` API and a migration plan |
| an import cycle | services that call each other in a loop |
| a `utils` package everyone imports | a shared library 30 services must upgrade |

Two services that must always be deployed together are tightly coupled, however many repositories they live in. That's sometimes called a *distributed monolith*: all the costs of a network, none of the independence. And unlike a function call, a call between services can fail halfway; see the *Distributed systems basics* guide.

## Common mistakes

- **One package per type, or one giant package.** Split by what changes together, not by file size.
- **`utils`, `common`, `helpers`.** Name packages after what they provide; split the drawer into real packages.
- **Exporting everything "just in case".** Every exported name is a promise you'll have to keep.
- **The core importing the details.** If your rollout logic imports your Slack package, flip it with an interface.
- **Fixing a cycle by merging everything.** Sometimes right, but first try moving the shared type down.
- **Adding a dependency for one small function,** or the reverse: hand-writing crypto to avoid one.
- **Not committing `go.sum`,** or editing it by hand.

## Check yourself

1. Package `helpers` contains `RetryHTTP`, `ParseCron`, `FormatBytes` and `SlackEscape`. What's wrong, and what would you do?
2. Your `report` package reads `aws.Fetcher.rawPages` directly. Why is this a problem, and what should `report` take instead?
3. You're designing an `inventory` package. You can't decide whether to export the helper `splitFields`. Which choice is safer, and why?
4. `alerts` imports `oncall` to find who to page, and `oncall` imports `alerts` to read `alerts.Severity`. What happens when you build? Give two ways to fix it.
5. A library goes from `v1.8.3` to `v1.9.0`. What does semantic versioning promise? What about `v2.0.0`?
6. You need to check whether a string is a valid Kubernetes object name (a DNS label). A popular library does it, plus 200 other validations, and pulls in 14 dependencies. What would you do?

## Answers

1. Low cohesion: four unrelated jobs in one drawer that everything imports. Split by purpose: retries next to the HTTP client code, cron parsing in a `schedule` package, Slack escaping inside `slack`.
2. That's tight coupling to internals: any change inside the AWS code breaks the report. Have `report` take plain data, like a `[]Host`, that the AWS code returns.
3. Leave it unexported. You can always export it later without breaking anyone; un-exporting it once others use it is a breaking change.
4. The build fails with `import cycle not allowed`. Move `Severity` into a small package both can import, or have `oncall` define an interface it needs instead of importing `alerts`. If the two are really one concept, merge them.
5. `v1.9.0` promises new features with nothing removed or changed incompatibly, so your code should keep compiling. `v2.0.0` may break your code, and in Go it has a different import path (`/v2`), so you must opt in.
6. Probably write it yourself: the rule (lowercase letters, digits and dashes, 1–63 characters, starting and ending with a letter or digit) is a few lines, plus a table test. You avoid 14 dependencies' worth of upgrades and supply-chain risk. "A little copying is better than a little dependency."

## Key terms

| Term | Meaning |
|---|---|
| module (general) | any part of a program with a boundary: function, package, service |
| Go module | a versioned collection of packages with a `go.mod` file |
| cohesion | how strongly the contents of one part belong together |
| coupling | how much one part depends on another's details |
| public surface | the names other code can use; everything else is internal |
| exported / unexported | capitalised names are visible outside the package; lowercase are not |
| dependency inversion | the core defines an interface; details implement it |
| dependency graph | packages as boxes, imports as arrows |
| cycle | packages that import each other, directly or through others |
| semantic versioning | `MAJOR.MINOR.PATCH`, where the bumped number signals what changed |
| minimal version selection | Go uses the highest version anyone asked for, never newer |
| supply-chain attack | getting malicious code into a program through its dependencies |
| distributed monolith | services so coupled they must be deployed together |

## Watch

- [Gopherfest 2015 | Go Proverbs with Rob Pike](https://www.youtube.com/watch?v=PAAkCSZUG1c) · The Go Programming Language · 22 min: Go's short design rules explained, including "a little copying is better than a little dependency".
- [GopherCon 2018: How Do You Structure Your Go Apps - Kat Zien](https://www.youtube.com/watch?v=oL6JBUk6tj0) · Gopher Academy · 46 min: several ways to split one app into packages, and how dependency direction keeps the core free of details.
- [Go Class: 40 Go Modules](https://www.youtube.com/watch?v=_N6BxmbLYBk) · Matt KØDVB · 18 min: `go.mod`, `go.sum`, semantic versioning and how Go picks dependency versions.
- [XZ Exploit - Computerphile](https://www.youtube.com/watch?v=LY8_35a_DEA) · Computerphile · 15 min: the story of the xz backdoor, a real supply-chain attack that nearly reached every Linux server.

## Go deeper

- [Go at Google: Language Design in the Service of Software Engineering](https://go.dev/talks/2012/splash.article) (Rob Pike): why Go handles imports, dependencies and cycles the way it does.
- [Organizing a Go module](https://go.dev/doc/modules/layout): the official guide to laying out packages, commands and `internal/`.
- [Our Software Dependency Problem](https://research.swtch.com/deps) (Russ Cox): how to judge a dependency before you add it, by the designer of Go modules.
- [How Go Mitigates Supply Chain Attacks](https://go.dev/blog/supply-chain): what `go.sum`, the checksum database and "building doesn't execute" protect you from.
- [Semantic Versioning 2.0.0](https://semver.org/): the short spec behind every `vX.Y.Z` tag.
