# Gopherlings lesson plan

<!-- begin summary -->
286 exercises in 14 modules, about 120 hours of hands-on work.
<!-- end summary -->
The course takes a complete beginner to the Go that DevOps and platform engineers write every day: CLIs, log and config processing, HTTP services, concurrency, Kubernetes controllers and AWS automation.

## How the course is built

Each **concept** has a short run of exercises, so one lesson is never your only exposure to it:

1. **The introduction** (for example *Variables*) teaches the idea from zero, with examples, and keeps its tasks small and guided.
2. **A "fix the bug" round** (*Variables: fix the broken builders*) shows the same idea from another angle, often as broken code or a compiler error to read.
3. **A small realistic job** (*Variables: build ops commands*) needs the idea with less scaffolding. Modules 1–5 have this third round; later modules have two exercises per concept.

Every module ends with a **practice** exercise that combines its concepts. Each phase ends with a multi-file **project**, where tests unlock task by task.

**Guides** teach the computer science underneath the exercises: one idea each, from *How a program runs* to *Concurrency* and *Declarative systems and control loops*. Each module lists its guides below. Read a module's guides before or alongside its exercises: press `g` in the watcher, run `gopherlings guides`, or open `guides/` in your course folder. Every guide ends with *Check yourself* questions; answer them out loud before reading the answers.

After you pass an exercise, press `s` to compare your code with the reference solution (also saved in `solutions/`). Reading a second way to solve the same problem is one of the fastest ways to learn.

## Suggested schedule

| Pace | Time per week | Finishes in about |
|---|---|---|
| Steady | 5 hours (an hour a weekday) | 6 months |
| Focused | 10 hours | 12 weeks |
| Bootcamp | 20 hours | 6 weeks |

Habits that make it stick:

- **Read the README first** (`i`). Each one teaches everything its tests check.
- **Use hints in order** (`h`). The first hint is a nudge and the last is close to the answer. Try for 5–10 minutes before each one.
- **Stuck after the last hint?** Press `s` for the solution, read it, then `gopherlings reset` the exercise the next day and solve it without looking. The list marks it ◐ until you do.
- **Space the repetitions.** Do a concept's introduction, then come back to its follow-ups the next day, before you've fully forgotten it.
- **Redo practice exercises cold.** After finishing a module, `gopherlings reset <practice>` and solve it again without hints.
- **Pair a guide with practice.** Read one section of a guide, then do an exercise. Short reading, then doing, beats a long reading session.
- **Explain it out loud.** After each module, say in two or three sentences what each concept is for. If you can't, revisit its introduction.

<!-- Generated from content/course.json; regenerate with:
     go test ./content -run TestLessonPlan -args -update -->
<!-- begin modules -->
## Phase: Go foundations (about 42 hours)

### Module 1: Basics

*About 6.0 hours, in `exercises/01_basics/`.* Functions, variables, constants, numbers, strings and fmt: the vocabulary every Go program is written in.

**After this module you can:** Read and write small Go functions: variables, constants, int/float64 maths, the strings package, fmt verbs and type conversion.

**Guides:** [How a program runs](content/guides/how_programs_run.md) · [Values and types](content/guides/values_and_types.md) · [Solving problems you haven't seen before](content/guides/problem_solving.md)

| # | Exercise | Kind | Min |
|---:|---|---|---:|
| 001 | Hello, Go | concept | 10 |
| 002 | Reading Go code: the parts of a file | concept | 20 |
| 003 | Variables | concept | 15 |
| 004 | Variables: fix the broken builders | concept | 10 |
| 005 | Variables: build ops commands | concept | 15 |
| 006 | Reading compiler errors | concept | 20 |
| 007 | Constants | concept | 15 |
| 008 | Constants: fix the health check config | concept | 10 |
| 009 | Constants: write your own config block | concept | 15 |
| 010 | Numbers and arithmetic | concept | 20 |
| 011 | Numbers: fix the metrics math | concept | 15 |
| 012 | Numbers: plan the capacity | concept | 15 |
| 013 | Strings | concept | 20 |
| 014 | Strings: fix the log and host checks | concept | 15 |
| 015 | Strings: clean up config and logs | concept | 15 |
| 016 | Formatting with fmt | concept | 20 |
| 017 | Formatting: fix the bot messages | concept | 15 |
| 018 | Formatting: build an ops report | concept | 15 |
| 019 | Type conversion | concept | 20 |
| 020 | Type conversion: fix the capacity helpers | concept | 15 |
| 021 | Type conversion: report disk usage | concept | 15 |
| 022 | Practice: deploy summary | **practice** | 30 |

### Module 2: Control flow and functions

*About 7.5 hours, in `exercises/02_control_flow/`.* Decisions, loops and functions: how a program chooses what to do and repeats work.

**After this module you can:** Make decisions and repeat work: booleans, if/else, switch, for loops, helper functions, (value, error) returns and closures. Parse a "host:port" safely.

**Guides:** [State and tracing code by hand](content/guides/tracing_state.md) · [Functions and decomposition](content/guides/functions_decomposition.md) · [Debugging with a method](content/guides/debugging.md)

| # | Exercise | Kind | Min |
|---:|---|---|---:|
| 023 | Booleans | concept | 15 |
| 024 | Booleans: fix the alert rules | concept | 10 |
| 025 | Booleans: autoscaler decisions | concept | 15 |
| 026 | if and else | concept | 20 |
| 027 | if and else: fix the maintenance script | concept | 20 |
| 028 | if and else: the deploy gate | concept | 20 |
| 029 | Zero values: default the config | concept | 15 |
| 030 | Zero values: fix the log rotation defaults | concept | 10 |
| 031 | Zero values: layered timeouts and probes | concept | 15 |
| 032 | Switch: size, method, latency | concept | 20 |
| 033 | Switch: the fallthrough bug | concept | 10 |
| 034 | Switch: container supervisor | concept | 20 |
| 035 | For loops: backoff and progress bars | concept | 20 |
| 036 | For loops: fix the off-by-one bugs | concept | 15 |
| 037 | For loops: plan a rollout | concept | 20 |
| 038 | Functions: a human-readable byte formatter | concept | 20 |
| 039 | Functions: swapped arguments and helpers | concept | 15 |
| 040 | Functions: SLO error budget report | concept | 20 |
| 041 | Multiple return values | concept | 20 |
| 042 | Multiple return values: don't drop the error | concept | 15 |
| 043 | Multiple return values: image refs and TTLs | concept | 20 |
| 044 | Functions as values | concept | 25 |
| 045 | Functions as values: state that sticks | concept | 15 |
| 046 | Functions as values: retries and a circuit breaker | concept | 25 |
| 047 | Practice: HTTP status triage | **practice** | 30 |

### Module 3: Collections

*About 9.3 hours, in `exercises/03_collections/`.* Arrays, slices, maps and strings as data: storing many values and working through them.

**After this module you can:** Store and process many values with arrays, slices, maps and sorting, avoid the shared-memory slice traps, and handle UTF-8 text correctly.

**Guides:** [Data structures](content/guides/data_structures.md) · [Algorithms and Big-O](content/guides/algorithms_big_o.md)

| # | Exercise | Kind | Min |
|---:|---|---|---:|
| 048 | Arrays | concept | 15 |
| 049 | Arrays: fix the hourly histogram | concept | 15 |
| 050 | Arrays: weekly on-call summary | concept | 15 |
| 051 | Slices | concept | 20 |
| 052 | Slices: fix the inventory script | concept | 15 |
| 053 | Slices: rolling-deploy batches | concept | 15 |
| 054 | append and copy | concept | 25 |
| 055 | append and copy: fix the inventory | concept | 15 |
| 056 | append and copy: rolling deploy plan | concept | 15 |
| 057 | Looping with range | concept | 20 |
| 058 | Looping with range: fix the alert math | concept | 15 |
| 059 | Looping with range: rollout report | concept | 15 |
| 060 | Variadic functions | concept | 15 |
| 061 | Variadic functions: fix the pager bot | concept | 15 |
| 062 | Variadic functions: a command builder | concept | 15 |
| 063 | Maps | concept | 25 |
| 064 | Maps: fix the config bugs | concept | 15 |
| 065 | Maps: diff a service catalog | concept | 15 |
| 066 | Sorting | concept | 25 |
| 067 | Sorting: fix the node report | concept | 15 |
| 068 | Sorting: the noisiest alerts | concept | 15 |
| 069 | The strings package | concept | 25 |
| 070 | The strings package: fix the bootstrap parser | concept | 15 |
| 071 | The strings package: which pods are broken? | concept | 15 |
| 072 | Bytes, runes and UTF-8 | concept | 25 |
| 073 | Bytes, runes and UTF-8: fix the byte bugs | concept | 15 |
| 074 | Bytes, runes and UTF-8: DNS-safe service names | concept | 15 |
| 075 | Slice gotchas: shared memory | concept | 30 |
| 076 | Slice gotchas: two bugs from one buffer | concept | 15 |
| 077 | Slice gotchas: drain and promote without side effects | concept | 15 |
| 078 | Practice: log level counter | **practice** | 35 |

### Module 4: Types and methods

*About 11.8 hours, in `exercises/04_types/`.* Structs, pointers, methods and interfaces: modelling real things like servers, deployments and alerts.

**After this module you can:** Model real things (deployments, hosts, alerts) with structs, pointers, methods, enums, interfaces, embedding and generics.

**Guides:** [Memory, pointers and sharing](content/guides/memory_pointers.md) · [Abstraction](content/guides/abstraction.md)

| # | Exercise | Kind | Min |
|---:|---|---|---:|
| 079 | Structs | concept | 25 |
| 080 | Structs: changes that vanish | concept | 15 |
| 081 | Structs: parse a host inventory | concept | 25 |
| 082 | Pointers | concept | 25 |
| 083 | Pointers: fix the silent no-ops | concept | 20 |
| 084 | Pointers: optional config fields | concept | 20 |
| 085 | Methods | concept | 20 |
| 086 | Methods: certificate checks | concept | 20 |
| 087 | Methods: host inventory | concept | 20 |
| 088 | Pointer receivers | concept | 25 |
| 089 | Pointer receivers: the leaky rate limiter | concept | 20 |
| 090 | Pointer receivers: a rolling deploy | concept | 20 |
| 091 | Custom types and enums | concept | 25 |
| 092 | Custom types: fix the deploy phase enum | concept | 15 |
| 093 | Custom types: Kubernetes memory quantities | concept | 25 |
| 094 | Interfaces | concept | 25 |
| 095 | Interfaces: fix the health checks | concept | 15 |
| 096 | Interfaces: swap in a fake for tests | concept | 20 |
| 097 | Stringers | concept | 20 |
| 098 | Stringers: pod phases and uptime | concept | 20 |
| 099 | Stringers: a kubectl-style deployment report | concept | 25 |
| 100 | Struct embedding | concept | 25 |
| 101 | Struct embedding: fix the Secret bugs | concept | 20 |
| 102 | Struct embedding: a shared health-check base | concept | 25 |
| 103 | Type assertions and switches | concept | 25 |
| 104 | Type switches: read decoded config values | concept | 20 |
| 105 | Type switches: flatten a nested config | concept | 25 |
| 106 | Generics | concept | 30 |
| 107 | Generics: Sum, SortedKeys and GroupBy | concept | 25 |
| 108 | Generics: a Ring of recent deploys | concept | 25 |
| 109 | Practice: service registry | **practice** | 40 |

### Module 5: Errors

*About 7.4 hours, in `exercises/05_errors/`.* Go treats errors as ordinary values. Learn to create, wrap, inspect and combine them the way production code does.

**After this module you can:** Handle failure the Go way: create, wrap, inspect and join errors, clean up with defer, and contain panics.

**Guides:** [Errors and failure](content/guides/errors_failure.md)

| # | Exercise | Kind | Min |
|---:|---|---|---:|
| 110 | Errors are values | concept | 20 |
| 111 | Errors are values: fix the replica scaler | concept | 15 |
| 112 | Errors are values: parse health checks | concept | 20 |
| 113 | Custom error types | concept | 20 |
| 114 | Custom error types: fix the port allocator | concept | 15 |
| 115 | Custom error types: retry failed health checks | concept | 20 |
| 116 | Wrapping errors | concept | 30 |
| 117 | Wrapping errors: fix the secrets fetcher | concept | 20 |
| 118 | Wrapping errors: triage a failed deploy | concept | 30 |
| 119 | Reporting many errors | concept | 25 |
| 120 | Reporting many errors: fix the firewall validator | concept | 20 |
| 121 | Reporting many errors: run every shutdown step | concept | 25 |
| 122 | defer | concept | 25 |
| 123 | defer: fix the log rotation job | concept | 20 |
| 124 | defer: safe database migrations | concept | 25 |
| 125 | panic and recover | concept | 25 |
| 126 | panic and recover: fix the chat bot's safety net | concept | 20 |
| 127 | panic and recover: protect the status API | concept | 30 |
| 128 | Practice: config validator | **practice** | 40 |

## Phase: Practical Go (about 19 hours)

### Module 6: Text and data

*About 8.2 hours, in `exercises/06_text_and_data/`.* Reading lines, regular expressions, time, JSON, CSV and templates: the daily bread of ops tooling.

**After this module you can:** Parse logs and configs: bufio scanning, regular expressions, time, JSON, CSV and text/template. The project produces an access-log report.

**Guides:** [Data formats and parsing](content/guides/data_formats.md)

| # | Exercise | Kind | Min |
|---:|---|---|---:|
| 129 | Reading input line by line: a hosts list | concept | 25 |
| 130 | Reading input line by line: scanning logs | concept | 20 |
| 131 | Regular expressions: release tags and log lines | concept | 30 |
| 132 | Regular expressions: fix the patterns | concept | 20 |
| 133 | Time and durations: cert expiry, uptime, maintenance windows | concept | 30 |
| 134 | Time and durations: fix the backup scheduler | concept | 20 |
| 135 | Encoding JSON: a deployment status report | concept | 30 |
| 136 | Encoding JSON: fix the alert stream | concept | 20 |
| 137 | Decoding JSON: load a service config | concept | 30 |
| 138 | Decoding JSON: fix the kubectl pod reader | concept | 20 |
| 139 | CSV files: a host inventory | concept | 25 |
| 140 | CSV files: fix the TSV DNS export | concept | 20 |
| 141 | Text templates: render an nginx upstream block | concept | 30 |
| 142 | Text templates: systemd unit files | concept | 20 |
| 143 | Practice: .env file parser | **practice** | 35 |
| 144 | Project: access log report | **project** | 120 |

### Module 7: Testing

*About 6.0 hours, in `exercises/07_testing/`.* Now you write the tests. Table tests, subtests, golden files, fakes, benchmarks and fuzzing.

**After this module you can:** Write your own tests: table tests, subtests, golden files, fakes, benchmarks and fuzzing. Your tests are graded by whether they catch real bugs.

**Guides:** [Testing and correctness](content/guides/testing_correctness.md)

| # | Exercise | Kind | Min |
|---:|---|---|---:|
| 145 | Your first test | concept | 20 |
| 146 | Your first test: fix a broken test | concept | 20 |
| 147 | Table-driven tests | concept | 25 |
| 148 | Table-driven tests: error cases | concept | 20 |
| 149 | Subtests and helpers | concept | 25 |
| 150 | Subtests and helpers: files and t.TempDir | concept | 20 |
| 151 | Test fixtures and golden files | concept | 30 |
| 152 | Golden files: write your own | concept | 20 |
| 153 | Fakes and dependency injection | concept | 30 |
| 154 | Fakes: a fake clock for retries | concept | 20 |
| 155 | Benchmarks | concept | 25 |
| 156 | Benchmarks: sub-benchmarks and preallocation | concept | 20 |
| 157 | Fuzzing | concept | 30 |
| 158 | Fuzzing: properties of a truncated alert | concept | 20 |
| 159 | Practice: reproduce a bug report | **practice** | 35 |

### Module 8: Packages and project layout

*About 4.2 hours, in `exercises/08_packages/`.* Split code into packages, control what is exported, embed files and stamp version info into binaries.

**After this module you can:** Split a program into packages, control your public API, use internal/, embed files and stamp version info into binaries.

**Guides:** [Modularity and dependencies](content/guides/modularity.md)

| # | Exercise | Kind | Min |
|---:|---|---|---:|
| 160 | Packages and imports | concept | 25 |
| 161 | Packages: two packages, one name | concept | 20 |
| 162 | Exported and unexported names | concept | 20 |
| 163 | Exported names: design a small API | concept | 20 |
| 164 | internal packages | concept | 20 |
| 165 | internal packages: the boundary | concept | 20 |
| 166 | Embedding files | concept | 25 |
| 167 | Embedding files: database migrations | concept | 20 |
| 168 | Version info and ldflags | concept | 25 |
| 169 | Build info: what's inside the binary | concept | 20 |
| 170 | Practice: lay out a tool | **practice** | 40 |

## Phase: Building tools and services (about 42 hours)

### Module 9: CLI tools and the OS

*About 12.1 hours, in `exercises/09_cli/`.* Arguments, flags, files, environment, subprocesses and signals: build the kind of tool you would put in a pipeline.

**After this module you can:** Build pipeline-ready CLIs: args, flags, subcommands, streams and exit codes, safe file writes, directory walks, config precedence, subprocesses, signals, checksums and archives.

**Guides:** [The operating system for programmers](content/guides/operating_systems.md)

| # | Exercise | Kind | Min |
|---:|---|---|---:|
| 171 | Command-line arguments: os.Args and usage errors | concept | 20 |
| 172 | Command-line arguments: fix the port waiter | concept | 20 |
| 173 | Flags: parsing options with a FlagSet | concept | 25 |
| 174 | Flags: repeatable and validated flags | concept | 25 |
| 175 | Subcommands: dispatching git-style commands | concept | 30 |
| 176 | Subcommands: a command table and generated help | concept | 25 |
| 177 | Standard streams: stdin, stdout, stderr and exit codes | concept | 25 |
| 178 | Standard streams: grep-style exit codes | concept | 20 |
| 179 | Files: read, scan and write | concept | 25 |
| 180 | Files: an append-only audit log | concept | 20 |
| 181 | Atomic writes: temp file and rename | concept | 30 |
| 182 | Atomic writes: safely change one setting | concept | 25 |
| 183 | Walking directories: find big log files | concept | 25 |
| 184 | Walking directories: fix the backup cleaner | concept | 25 |
| 185 | Config precedence: defaults, file, env and flags | concept | 30 |
| 186 | Config precedence: fix the backup job config | concept | 30 |
| 187 | Running commands: output, exit codes and timeouts | concept | 30 |
| 188 | Running commands: stdin, stderr and PATH lookups | concept | 30 |
| 189 | Signals: graceful stop for a rolling drain | concept | 30 |
| 190 | Signals: SIGHUP reloads, SIGTERM stops | concept | 20 |
| 191 | Checksums and archives: verify and extract safely | concept | 35 |
| 192 | Checksums and archives: publish a release | concept | 30 |
| 193 | Project: ops inventory CLI | **project** | 150 |

### Module 10: Networking and HTTP

*About 9.1 hours, in `exercises/10_networking/`.* IP addresses, TCP, HTTP clients and servers, JSON APIs, middleware and TLS.

**After this module you can:** Work with IPs and CIDRs, TCP, port checks, HTTP clients and servers, JSON APIs, middleware and TLS, all tested with httptest.

**Guides:** [Networking from the ground up](content/guides/networking.md)

| # | Exercise | Kind | Min |
|---:|---|---|---:|
| 194 | IP addresses: classify, allowlist, overlap | concept | 25 |
| 195 | CIDRs: normalize and audit subnet rules | concept | 20 |
| 196 | TCP: a line-based control protocol | concept | 30 |
| 197 | TCP: banner grabbing with deadlines | concept | 25 |
| 198 | Port checks: open, closed or filtered | concept | 30 |
| 199 | Port checks: wait for the database | concept | 25 |
| 200 | HTTP clients: fetch a service status | concept | 30 |
| 201 | HTTP clients: fix the deploy notifier | concept | 25 |
| 202 | HTTP servers: handlers and routes | concept | 30 |
| 203 | HTTP servers: fix the log viewer | concept | 30 |
| 204 | Testing HTTP: recorders and fake servers | concept | 25 |
| 205 | Testing HTTP: a regression test for a bug report | concept | 25 |
| 206 | JSON APIs: build a service catalog | concept | 35 |
| 207 | JSON APIs: fix the alerts API | concept | 30 |
| 208 | Middleware: logging, request IDs and panics | concept | 30 |
| 209 | Middleware: chains, status codes and auth | concept | 30 |
| 210 | TLS: trust a CA and read certificates | concept | 30 |
| 211 | TLS: stop skipping verification | concept | 30 |
| 212 | Practice: endpoint health checker | **practice** | 40 |

### Module 11: Concurrency

*About 10.5 hours, in `exercises/11_concurrency/`.* Goroutines, channels, select, sync and context. Every check runs with the race detector.

**After this module you can:** Use goroutines, channels, select, WaitGroups, mutexes and context; read race reports; build worker pools and pipelines; collect errors concurrently.

**Guides:** [Concurrency](content/guides/concurrency.md)

| # | Exercise | Kind | Min |
|---:|---|---|---:|
| 213 | Goroutines | concept · race | 20 |
| 214 | Goroutines: fix the maintenance script | concept · race | 15 |
| 215 | Channels | concept · race | 25 |
| 216 | Channels: deadlocks and directions | concept · race | 20 |
| 217 | select and timeouts | concept · race | 25 |
| 218 | select: tickers and deadlines | concept · race | 20 |
| 219 | WaitGroups | concept · race | 20 |
| 220 | WaitGroups: patching the fleet | concept · race | 20 |
| 221 | Mutexes | concept · race | 25 |
| 222 | Mutexes: atomics and slow locks | concept · race | 20 |
| 223 | The race detector | concept · race | 25 |
| 224 | The race detector: hidden races | concept · race | 20 |
| 225 | Context and cancellation | concept · race | 30 |
| 226 | Context: cancelling HTTP requests | concept · race | 25 |
| 227 | Worker pools | concept · race | 30 |
| 228 | Worker pools: fix the image puller | concept · race | 20 |
| 229 | Pipelines | concept · race | 30 |
| 230 | Pipelines: fan-in and cancellation | concept · race | 30 |
| 231 | Collecting errors from goroutines | concept · race | 30 |
| 232 | Errors from goroutines: stop on first failure | concept · race | 30 |
| 233 | Project: uptime monitor | **project** · race | 150 |

### Module 12: Production services

*About 10.8 hours, in `exercises/12_production/`.* Logging, metrics, health checks, graceful shutdown, retries, rate limits and caches: what keeps a service up at 3am.

**After this module you can:** Add what on-call needs: slog logging, Prometheus metrics, health checks, graceful shutdown, retries, rate limits, circuit breakers, caches and idempotency.

**Guides:** [Distributed systems basics](content/guides/distributed_systems.md)

| # | Exercise | Kind | Min |
|---:|---|---|---:|
| 234 | Structured logging with slog | concept · race | 25 |
| 235 | Structured logging: access logs | concept · race | 20 |
| 236 | Prometheus-style metrics: counters and /metrics | concept · race | 30 |
| 237 | Prometheus-style metrics: histograms | concept · race | 25 |
| 238 | Health checks: liveness and readiness | concept · race | 25 |
| 239 | Health checks: readiness gates | concept · race | 20 |
| 240 | Graceful shutdown: drain in-flight requests | concept · race | 30 |
| 241 | Graceful shutdown: HTTP server and worker | concept · race | 20 |
| 242 | Retries: backoff and jitter | concept · race | 30 |
| 243 | Retries: retryable HTTP errors | concept · race | 25 |
| 244 | Rate limiting: token buckets | concept · race | 30 |
| 245 | Rate limiting: fix the bucket | concept · race | 25 |
| 246 | Circuit breakers: fail fast | concept · race | 30 |
| 247 | Circuit breakers: protect an HTTP client | concept · race | 25 |
| 248 | TTL caching: expire on read | concept · race | 30 |
| 249 | TTL caching: collapse concurrent loads | concept · race | 25 |
| 250 | Idempotency keys: replay retried requests | concept · race | 30 |
| 251 | Idempotency keys: scope, expire and fingerprint | concept · race | 20 |
| 252 | Project: inventory service | **project** · race | 180 |

## Phase: Platform engineering (about 17 hours)

### Module 13: Kubernetes with client-go

*About 10.9 hours, in `exercises/13_kubernetes/`.* Kubernetes objects, manifests, kubeconfig, clientsets, informers, workqueues and reconcile loops using the real client-go.

**After this module you can:** Build and validate Kubernetes objects, read manifests and kubeconfigs, use the clientset, selectors, watches, informers and workqueues, and write a controller.

**Guides:** [Declarative systems and control loops](content/guides/control_loops.md)

| # | Exercise | Kind | Min |
|---:|---|---|---:|
| 253 | Kubernetes objects: building a Deployment | concept | 30 |
| 254 | Kubernetes objects: Services and selectors | concept | 25 |
| 255 | YAML manifests: reading them into Go | concept | 30 |
| 256 | YAML manifests: decoding kinds and linting images | concept | 30 |
| 257 | kubeconfig: clusters, users and contexts | concept | 25 |
| 258 | kubeconfig: building a rest.Config | concept | 20 |
| 259 | The clientset: get, create and delete | concept | 30 |
| 260 | The clientset: updates and create-or-update | concept | 30 |
| 261 | Labels and selectors: matching in Go | concept | 25 |
| 262 | Labels and selectors: selecting on the server | concept | 25 |
| 263 | Watching resources: streaming Pod events | concept · race | 30 |
| 264 | Watching resources: closed watches and ERROR events | concept · race | 25 |
| 265 | Informers: list, watch and cache | concept · race | 35 |
| 266 | Informers: listers and the read-only cache | concept · race | 30 |
| 267 | Workqueues: keys, dedup and the worker loop | concept · race | 30 |
| 268 | Workqueues: retries with backoff | concept · race | 25 |
| 269 | Reconcile loops: desired vs actual | concept | 30 |
| 270 | Reconcile loops: idempotent updates | concept | 30 |
| 271 | Project: Kubernetes controller | **project** · race | 150 |

### Module 14: Cloud automation with the AWS SDK

*About 6.4 hours, in `exercises/14_aws/`.* aws-sdk-go-v2 configuration, testable clients, paginators, API errors, S3 and tag compliance, all tested offline.

**After this module you can:** Configure aws-sdk-go-v2 safely, write testable SDK code, paginate, handle API errors, manage S3 objects, audit tags and detect drift.

**Guides:** [Security basics for automation](content/guides/security.md)

| # | Exercise | Kind | Min |
|---:|---|---|---:|
| 272 | AWS SDK configuration: regions, profiles and credentials | concept | 25 |
| 273 | AWS SDK configuration: many regions, no leaked keys | concept | 20 |
| 274 | SDK clients: small interfaces and fakes | concept | 30 |
| 275 | SDK clients: spy on StopInstances | concept | 20 |
| 276 | Paginators: walk every page | concept | 25 |
| 277 | Paginators: fix the stuck size report | concept | 20 |
| 278 | AWS API errors: codes, faults and NoSuchKey | concept | 25 |
| 279 | AWS API errors: EC2 dry-run permission checks | concept | 20 |
| 280 | S3 objects: upload, verify and list configs | concept | 30 |
| 281 | S3 objects: fix the log cleanup | concept | 20 |
| 282 | Tag compliance: audit and fix instance tags | concept | 30 |
| 283 | Tag compliance: untagged S3 buckets and bad values | concept | 20 |
| 284 | Drift detection: a tiny terraform plan | concept | 35 |
| 285 | Drift detection: ignore AWS-managed tags | concept | 20 |
| 286 | Practice: cloud inventory report | **practice** · race | 45 |

<!-- end modules -->

## After the course

Some next steps that build on what you've done:

- Turn the *ops inventory CLI* or *uptime monitor* project into a real tool you use at work. Add a `Makefile`, a GitHub Actions workflow and a `goreleaser` config.
- Rebuild the *Kubernetes controller* with [controller-runtime](https://book.kubebuilder.io/) and deploy it to a [kind](https://kind.sigs.k8s.io/) cluster.
- Read [Effective Go](https://go.dev/doc/effective_go) and [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments). Most of it will now feel familiar.
- Practise on [Exercism's Go track](https://exercism.org/tracks/go) to keep your fundamentals sharp.
