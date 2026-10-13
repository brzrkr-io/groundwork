# Gopherlings lesson plan

<!-- begin summary -->
284 exercises in 14 modules, about 120 hours of hands-on work.
<!-- end summary -->
The course takes a complete beginner to the Go that DevOps and platform engineers write every day: CLIs, log and config processing, HTTP services, concurrency, Kubernetes controllers and AWS automation.

## How the course is built

Each **concept** has a short run of exercises, so one lesson is never your only exposure to it:

1. **The introduction** (for example *Variables*) teaches the idea from zero, with examples, and keeps its tasks small and guided.
2. **A "fix the bug" round** (*Variables: fix the broken builders*) shows the same idea from another angle, often as broken code or a compiler error to read.
3. **A small realistic job** (*Variables: build ops commands*) needs the idea with less scaffolding. Modules 1–5 have this third round; later modules have two exercises per concept.

Every module ends with a **practice** exercise that combines its concepts. Each phase ends with a multi-file **project**, where tests unlock task by task.

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
- **Explain it out loud.** After each module, say in two or three sentences what each concept is for. If you can't, revisit its introduction.

<!-- Generated from content/course.json; regenerate with:
     go test ./content -run TestLessonPlan -args -update -->
<!-- begin modules -->
## Phase: Go foundations (about 41 hours)

### Module 1: Basics

*About 5.3 hours, in `exercises/01_basics/`.* Functions, variables, constants, numbers, strings and fmt: the vocabulary every Go program is written in.

**After this module you can:** Read and write small Go functions: variables, constants, int/float64 maths, the strings package, fmt verbs and type conversion.

| # | Exercise | Kind | Min |
|---:|---|---|---:|
| 001 | Hello, Go | concept | 10 |
| 002 | Variables | concept | 15 |
| 003 | Variables: fix the broken builders | concept | 10 |
| 004 | Variables: build ops commands | concept | 15 |
| 005 | Constants | concept | 15 |
| 006 | Constants: fix the health check config | concept | 10 |
| 007 | Constants: write your own config block | concept | 15 |
| 008 | Numbers and arithmetic | concept | 20 |
| 009 | Numbers: fix the metrics math | concept | 15 |
| 010 | Numbers: plan the capacity | concept | 15 |
| 011 | Strings | concept | 20 |
| 012 | Strings: fix the log and host checks | concept | 15 |
| 013 | Strings: clean up config and logs | concept | 15 |
| 014 | Formatting with fmt | concept | 20 |
| 015 | Formatting: fix the bot messages | concept | 15 |
| 016 | Formatting: build an ops report | concept | 15 |
| 017 | Type conversion | concept | 20 |
| 018 | Type conversion: fix the capacity helpers | concept | 15 |
| 019 | Type conversion: report disk usage | concept | 15 |
| 020 | Practice: deploy summary | **practice** | 30 |

### Module 2: Control flow and functions

*About 7.5 hours, in `exercises/02_control_flow/`.* Decisions, loops and functions: how a program chooses what to do and repeats work.

**After this module you can:** Make decisions and repeat work: booleans, if/else, switch, for loops, helper functions, (value, error) returns and closures. Parse a "host:port" safely.

| # | Exercise | Kind | Min |
|---:|---|---|---:|
| 021 | Booleans | concept | 15 |
| 022 | Booleans: fix the alert rules | concept | 10 |
| 023 | Booleans: autoscaler decisions | concept | 15 |
| 024 | if and else | concept | 20 |
| 025 | if and else: fix the maintenance script | concept | 20 |
| 026 | if and else: the deploy gate | concept | 20 |
| 027 | Zero values: default the config | concept | 15 |
| 028 | Zero values: fix the log rotation defaults | concept | 10 |
| 029 | Zero values: layered timeouts and probes | concept | 15 |
| 030 | Switch: size, method, latency | concept | 20 |
| 031 | Switch: the fallthrough bug | concept | 10 |
| 032 | Switch: container supervisor | concept | 20 |
| 033 | For loops: backoff and progress bars | concept | 20 |
| 034 | For loops: fix the off-by-one bugs | concept | 15 |
| 035 | For loops: plan a rollout | concept | 20 |
| 036 | Functions: a human-readable byte formatter | concept | 20 |
| 037 | Functions: swapped arguments and helpers | concept | 15 |
| 038 | Functions: SLO error budget report | concept | 20 |
| 039 | Multiple return values | concept | 20 |
| 040 | Multiple return values: don't drop the error | concept | 15 |
| 041 | Multiple return values: image refs and TTLs | concept | 20 |
| 042 | Functions as values | concept | 25 |
| 043 | Functions as values: state that sticks | concept | 15 |
| 044 | Functions as values: retries and a circuit breaker | concept | 25 |
| 045 | Practice: HTTP status triage | **practice** | 30 |

### Module 3: Collections

*About 9.3 hours, in `exercises/03_collections/`.* Arrays, slices, maps and strings as data: storing many values and working through them.

**After this module you can:** Store and process many values with arrays, slices, maps and sorting, avoid the shared-memory slice traps, and handle UTF-8 text correctly.

| # | Exercise | Kind | Min |
|---:|---|---|---:|
| 046 | Arrays | concept | 15 |
| 047 | Arrays: fix the hourly histogram | concept | 15 |
| 048 | Arrays: weekly on-call summary | concept | 15 |
| 049 | Slices | concept | 20 |
| 050 | Slices: fix the inventory script | concept | 15 |
| 051 | Slices: rolling-deploy batches | concept | 15 |
| 052 | append and copy | concept | 25 |
| 053 | append and copy: fix the inventory | concept | 15 |
| 054 | append and copy: rolling deploy plan | concept | 15 |
| 055 | Looping with range | concept | 20 |
| 056 | Looping with range: fix the alert math | concept | 15 |
| 057 | Looping with range: rollout report | concept | 15 |
| 058 | Variadic functions | concept | 15 |
| 059 | Variadic functions: fix the pager bot | concept | 15 |
| 060 | Variadic functions: a command builder | concept | 15 |
| 061 | Maps | concept | 25 |
| 062 | Maps: fix the config bugs | concept | 15 |
| 063 | Maps: diff a service catalog | concept | 15 |
| 064 | Sorting | concept | 25 |
| 065 | Sorting: fix the node report | concept | 15 |
| 066 | Sorting: the noisiest alerts | concept | 15 |
| 067 | The strings package | concept | 25 |
| 068 | The strings package: fix the bootstrap parser | concept | 15 |
| 069 | The strings package: which pods are broken? | concept | 15 |
| 070 | Bytes, runes and UTF-8 | concept | 25 |
| 071 | Bytes, runes and UTF-8: fix the byte bugs | concept | 15 |
| 072 | Bytes, runes and UTF-8: DNS-safe service names | concept | 15 |
| 073 | Slice gotchas: shared memory | concept | 30 |
| 074 | Slice gotchas: two bugs from one buffer | concept | 15 |
| 075 | Slice gotchas: drain and promote without side effects | concept | 15 |
| 076 | Practice: log level counter | **practice** | 35 |

### Module 4: Types and methods

*About 11.8 hours, in `exercises/04_types/`.* Structs, pointers, methods and interfaces: modelling real things like servers, deployments and alerts.

**After this module you can:** Model real things (deployments, hosts, alerts) with structs, pointers, methods, enums, interfaces, embedding and generics.

| # | Exercise | Kind | Min |
|---:|---|---|---:|
| 077 | Structs | concept | 25 |
| 078 | Structs: changes that vanish | concept | 15 |
| 079 | Structs: parse a host inventory | concept | 25 |
| 080 | Pointers | concept | 25 |
| 081 | Pointers: fix the silent no-ops | concept | 20 |
| 082 | Pointers: optional config fields | concept | 20 |
| 083 | Methods | concept | 20 |
| 084 | Methods: certificate checks | concept | 20 |
| 085 | Methods: host inventory | concept | 20 |
| 086 | Pointer receivers | concept | 25 |
| 087 | Pointer receivers: the leaky rate limiter | concept | 20 |
| 088 | Pointer receivers: a rolling deploy | concept | 20 |
| 089 | Custom types and enums | concept | 25 |
| 090 | Custom types: fix the deploy phase enum | concept | 15 |
| 091 | Custom types: Kubernetes memory quantities | concept | 25 |
| 092 | Interfaces | concept | 25 |
| 093 | Interfaces: fix the health checks | concept | 15 |
| 094 | Interfaces: swap in a fake for tests | concept | 20 |
| 095 | Stringers | concept | 20 |
| 096 | Stringers: pod phases and uptime | concept | 20 |
| 097 | Stringers: a kubectl-style deployment report | concept | 25 |
| 098 | Struct embedding | concept | 25 |
| 099 | Struct embedding: fix the Secret bugs | concept | 20 |
| 100 | Struct embedding: a shared health-check base | concept | 25 |
| 101 | Type assertions and switches | concept | 25 |
| 102 | Type switches: read decoded config values | concept | 20 |
| 103 | Type switches: flatten a nested config | concept | 25 |
| 104 | Generics | concept | 30 |
| 105 | Generics: Sum, SortedKeys and GroupBy | concept | 25 |
| 106 | Generics: a Ring of recent deploys | concept | 25 |
| 107 | Practice: service registry | **practice** | 40 |

### Module 5: Errors

*About 7.4 hours, in `exercises/05_errors/`.* Go treats errors as ordinary values. Learn to create, wrap, inspect and combine them the way production code does.

**After this module you can:** Handle failure the Go way: create, wrap, inspect and join errors, clean up with defer, and contain panics.

| # | Exercise | Kind | Min |
|---:|---|---|---:|
| 108 | Errors are values | concept | 20 |
| 109 | Errors are values: fix the replica scaler | concept | 15 |
| 110 | Errors are values: parse health checks | concept | 20 |
| 111 | Custom error types | concept | 20 |
| 112 | Custom error types: fix the port allocator | concept | 15 |
| 113 | Custom error types: retry failed health checks | concept | 20 |
| 114 | Wrapping errors | concept | 30 |
| 115 | Wrapping errors: fix the secrets fetcher | concept | 20 |
| 116 | Wrapping errors: triage a failed deploy | concept | 30 |
| 117 | Reporting many errors | concept | 25 |
| 118 | Reporting many errors: fix the firewall validator | concept | 20 |
| 119 | Reporting many errors: run every shutdown step | concept | 25 |
| 120 | defer | concept | 25 |
| 121 | defer: fix the log rotation job | concept | 20 |
| 122 | defer: safe database migrations | concept | 25 |
| 123 | panic and recover | concept | 25 |
| 124 | panic and recover: fix the chat bot's safety net | concept | 20 |
| 125 | panic and recover: protect the status API | concept | 30 |
| 126 | Practice: config validator | **practice** | 40 |

## Phase: Practical Go (about 19 hours)

### Module 6: Text and data

*About 8.2 hours, in `exercises/06_text_and_data/`.* Reading lines, regular expressions, time, JSON, CSV and templates: the daily bread of ops tooling.

**After this module you can:** Parse logs and configs: bufio scanning, regular expressions, time, JSON, CSV and text/template. The project produces an access-log report.

| # | Exercise | Kind | Min |
|---:|---|---|---:|
| 127 | Reading input line by line: a hosts list | concept | 25 |
| 128 | Reading input line by line: scanning logs | concept | 20 |
| 129 | Regular expressions: release tags and log lines | concept | 30 |
| 130 | Regular expressions: fix the patterns | concept | 20 |
| 131 | Time and durations: cert expiry, uptime, maintenance windows | concept | 30 |
| 132 | Time and durations: fix the backup scheduler | concept | 20 |
| 133 | Encoding JSON: a deployment status report | concept | 30 |
| 134 | Encoding JSON: fix the alert stream | concept | 20 |
| 135 | Decoding JSON: load a service config | concept | 30 |
| 136 | Decoding JSON: fix the kubectl pod reader | concept | 20 |
| 137 | CSV files: a host inventory | concept | 25 |
| 138 | CSV files: fix the TSV DNS export | concept | 20 |
| 139 | Text templates: render an nginx upstream block | concept | 30 |
| 140 | Text templates: systemd unit files | concept | 20 |
| 141 | Practice: .env file parser | **practice** | 35 |
| 142 | Project: access log report | **project** | 120 |

### Module 7: Testing

*About 6.0 hours, in `exercises/07_testing/`.* Now you write the tests. Table tests, subtests, golden files, fakes, benchmarks and fuzzing.

**After this module you can:** Write your own tests: table tests, subtests, golden files, fakes, benchmarks and fuzzing. Your tests are graded by whether they catch real bugs.

| # | Exercise | Kind | Min |
|---:|---|---|---:|
| 143 | Your first test | concept | 20 |
| 144 | Your first test: fix a broken test | concept | 20 |
| 145 | Table-driven tests | concept | 25 |
| 146 | Table-driven tests: error cases | concept | 20 |
| 147 | Subtests and helpers | concept | 25 |
| 148 | Subtests and helpers: files and t.TempDir | concept | 20 |
| 149 | Test fixtures and golden files | concept | 30 |
| 150 | Golden files: write your own | concept | 20 |
| 151 | Fakes and dependency injection | concept | 30 |
| 152 | Fakes: a fake clock for retries | concept | 20 |
| 153 | Benchmarks | concept | 25 |
| 154 | Benchmarks: sub-benchmarks and preallocation | concept | 20 |
| 155 | Fuzzing | concept | 30 |
| 156 | Fuzzing: properties of a truncated alert | concept | 20 |
| 157 | Practice: reproduce a bug report | **practice** | 35 |

### Module 8: Packages and project layout

*About 4.2 hours, in `exercises/08_packages/`.* Split code into packages, control what is exported, embed files and stamp version info into binaries.

**After this module you can:** Split a program into packages, control your public API, use internal/, embed files and stamp version info into binaries.

| # | Exercise | Kind | Min |
|---:|---|---|---:|
| 158 | Packages and imports | concept | 25 |
| 159 | Packages: two packages, one name | concept | 20 |
| 160 | Exported and unexported names | concept | 20 |
| 161 | Exported names: design a small API | concept | 20 |
| 162 | internal packages | concept | 20 |
| 163 | internal packages: the boundary | concept | 20 |
| 164 | Embedding files | concept | 25 |
| 165 | Embedding files: database migrations | concept | 20 |
| 166 | Version info and ldflags | concept | 25 |
| 167 | Build info: what's inside the binary | concept | 20 |
| 168 | Practice: lay out a tool | **practice** | 40 |

## Phase: Building tools and services (about 42 hours)

### Module 9: CLI tools and the OS

*About 12.1 hours, in `exercises/09_cli/`.* Arguments, flags, files, environment, subprocesses and signals: build the kind of tool you would put in a pipeline.

**After this module you can:** Build pipeline-ready CLIs: args, flags, subcommands, streams and exit codes, safe file writes, directory walks, config precedence, subprocesses, signals, checksums and archives.

| # | Exercise | Kind | Min |
|---:|---|---|---:|
| 169 | Command-line arguments: os.Args and usage errors | concept | 20 |
| 170 | Command-line arguments: fix the port waiter | concept | 20 |
| 171 | Flags: parsing options with a FlagSet | concept | 25 |
| 172 | Flags: repeatable and validated flags | concept | 25 |
| 173 | Subcommands: dispatching git-style commands | concept | 30 |
| 174 | Subcommands: a command table and generated help | concept | 25 |
| 175 | Standard streams: stdin, stdout, stderr and exit codes | concept | 25 |
| 176 | Standard streams: grep-style exit codes | concept | 20 |
| 177 | Files: read, scan and write | concept | 25 |
| 178 | Files: an append-only audit log | concept | 20 |
| 179 | Atomic writes: temp file and rename | concept | 30 |
| 180 | Atomic writes: safely change one setting | concept | 25 |
| 181 | Walking directories: find big log files | concept | 25 |
| 182 | Walking directories: fix the backup cleaner | concept | 25 |
| 183 | Config precedence: defaults, file, env and flags | concept | 30 |
| 184 | Config precedence: fix the backup job config | concept | 30 |
| 185 | Running commands: output, exit codes and timeouts | concept | 30 |
| 186 | Running commands: stdin, stderr and PATH lookups | concept | 30 |
| 187 | Signals: graceful stop for a rolling drain | concept | 30 |
| 188 | Signals: SIGHUP reloads, SIGTERM stops | concept | 20 |
| 189 | Checksums and archives: verify and extract safely | concept | 35 |
| 190 | Checksums and archives: publish a release | concept | 30 |
| 191 | Project: ops inventory CLI | **project** | 150 |

### Module 10: Networking and HTTP

*About 9.1 hours, in `exercises/10_networking/`.* IP addresses, TCP, HTTP clients and servers, JSON APIs, middleware and TLS.

**After this module you can:** Work with IPs and CIDRs, TCP, port checks, HTTP clients and servers, JSON APIs, middleware and TLS, all tested with httptest.

| # | Exercise | Kind | Min |
|---:|---|---|---:|
| 192 | IP addresses: classify, allowlist, overlap | concept | 25 |
| 193 | CIDRs: normalize and audit subnet rules | concept | 20 |
| 194 | TCP: a line-based control protocol | concept | 30 |
| 195 | TCP: banner grabbing with deadlines | concept | 25 |
| 196 | Port checks: open, closed or filtered | concept | 30 |
| 197 | Port checks: wait for the database | concept | 25 |
| 198 | HTTP clients: fetch a service status | concept | 30 |
| 199 | HTTP clients: fix the deploy notifier | concept | 25 |
| 200 | HTTP servers: handlers and routes | concept | 30 |
| 201 | HTTP servers: fix the log viewer | concept | 30 |
| 202 | Testing HTTP: recorders and fake servers | concept | 25 |
| 203 | Testing HTTP: a regression test for a bug report | concept | 25 |
| 204 | JSON APIs: build a service catalog | concept | 35 |
| 205 | JSON APIs: fix the alerts API | concept | 30 |
| 206 | Middleware: logging, request IDs and panics | concept | 30 |
| 207 | Middleware: chains, status codes and auth | concept | 30 |
| 208 | TLS: trust a CA and read certificates | concept | 30 |
| 209 | TLS: stop skipping verification | concept | 30 |
| 210 | Practice: endpoint health checker | **practice** | 40 |

### Module 11: Concurrency

*About 10.5 hours, in `exercises/11_concurrency/`.* Goroutines, channels, select, sync and context. Every check runs with the race detector.

**After this module you can:** Use goroutines, channels, select, WaitGroups, mutexes and context; read race reports; build worker pools and pipelines; collect errors concurrently.

| # | Exercise | Kind | Min |
|---:|---|---|---:|
| 211 | Goroutines | concept · race | 20 |
| 212 | Goroutines: fix the maintenance script | concept · race | 15 |
| 213 | Channels | concept · race | 25 |
| 214 | Channels: deadlocks and directions | concept · race | 20 |
| 215 | select and timeouts | concept · race | 25 |
| 216 | select: tickers and deadlines | concept · race | 20 |
| 217 | WaitGroups | concept · race | 20 |
| 218 | WaitGroups: patching the fleet | concept · race | 20 |
| 219 | Mutexes | concept · race | 25 |
| 220 | Mutexes: atomics and slow locks | concept · race | 20 |
| 221 | The race detector | concept · race | 25 |
| 222 | The race detector: hidden races | concept · race | 20 |
| 223 | Context and cancellation | concept · race | 30 |
| 224 | Context: cancelling HTTP requests | concept · race | 25 |
| 225 | Worker pools | concept · race | 30 |
| 226 | Worker pools: fix the image puller | concept · race | 20 |
| 227 | Pipelines | concept · race | 30 |
| 228 | Pipelines: fan-in and cancellation | concept · race | 30 |
| 229 | Collecting errors from goroutines | concept · race | 30 |
| 230 | Errors from goroutines: stop on first failure | concept · race | 30 |
| 231 | Project: uptime monitor | **project** · race | 150 |

### Module 12: Production services

*About 10.8 hours, in `exercises/12_production/`.* Logging, metrics, health checks, graceful shutdown, retries, rate limits and caches: what keeps a service up at 3am.

**After this module you can:** Add what on-call needs: slog logging, Prometheus metrics, health checks, graceful shutdown, retries, rate limits, circuit breakers, caches and idempotency.

| # | Exercise | Kind | Min |
|---:|---|---|---:|
| 232 | Structured logging with slog | concept · race | 25 |
| 233 | Structured logging: access logs | concept · race | 20 |
| 234 | Prometheus-style metrics: counters and /metrics | concept · race | 30 |
| 235 | Prometheus-style metrics: histograms | concept · race | 25 |
| 236 | Health checks: liveness and readiness | concept · race | 25 |
| 237 | Health checks: readiness gates | concept · race | 20 |
| 238 | Graceful shutdown: drain in-flight requests | concept · race | 30 |
| 239 | Graceful shutdown: HTTP server and worker | concept · race | 20 |
| 240 | Retries: backoff and jitter | concept · race | 30 |
| 241 | Retries: retryable HTTP errors | concept · race | 25 |
| 242 | Rate limiting: token buckets | concept · race | 30 |
| 243 | Rate limiting: fix the bucket | concept · race | 25 |
| 244 | Circuit breakers: fail fast | concept · race | 30 |
| 245 | Circuit breakers: protect an HTTP client | concept · race | 25 |
| 246 | TTL caching: expire on read | concept · race | 30 |
| 247 | TTL caching: collapse concurrent loads | concept · race | 25 |
| 248 | Idempotency keys: replay retried requests | concept · race | 30 |
| 249 | Idempotency keys: scope, expire and fingerprint | concept · race | 20 |
| 250 | Project: inventory service | **project** · race | 180 |

## Phase: Platform engineering (about 17 hours)

### Module 13: Kubernetes with client-go

*About 10.9 hours, in `exercises/13_kubernetes/`.* Kubernetes objects, manifests, kubeconfig, clientsets, informers, workqueues and reconcile loops using the real client-go.

**After this module you can:** Build and validate Kubernetes objects, read manifests and kubeconfigs, use the clientset, selectors, watches, informers and workqueues, and write a controller.

| # | Exercise | Kind | Min |
|---:|---|---|---:|
| 251 | Kubernetes objects: building a Deployment | concept | 30 |
| 252 | Kubernetes objects: Services and selectors | concept | 25 |
| 253 | YAML manifests: reading them into Go | concept | 30 |
| 254 | YAML manifests: decoding kinds and linting images | concept | 30 |
| 255 | kubeconfig: clusters, users and contexts | concept | 25 |
| 256 | kubeconfig: building a rest.Config | concept | 20 |
| 257 | The clientset: get, create and delete | concept | 30 |
| 258 | The clientset: updates and create-or-update | concept | 30 |
| 259 | Labels and selectors: matching in Go | concept | 25 |
| 260 | Labels and selectors: selecting on the server | concept | 25 |
| 261 | Watching resources: streaming Pod events | concept · race | 30 |
| 262 | Watching resources: closed watches and ERROR events | concept · race | 25 |
| 263 | Informers: list, watch and cache | concept · race | 35 |
| 264 | Informers: listers and the read-only cache | concept · race | 30 |
| 265 | Workqueues: keys, dedup and the worker loop | concept · race | 30 |
| 266 | Workqueues: retries with backoff | concept · race | 25 |
| 267 | Reconcile loops: desired vs actual | concept | 30 |
| 268 | Reconcile loops: idempotent updates | concept | 30 |
| 269 | Project: Kubernetes controller | **project** · race | 150 |

### Module 14: Cloud automation with the AWS SDK

*About 6.4 hours, in `exercises/14_aws/`.* aws-sdk-go-v2 configuration, testable clients, paginators, API errors, S3 and tag compliance, all tested offline.

**After this module you can:** Configure aws-sdk-go-v2 safely, write testable SDK code, paginate, handle API errors, manage S3 objects, audit tags and detect drift.

| # | Exercise | Kind | Min |
|---:|---|---|---:|
| 270 | AWS SDK configuration: regions, profiles and credentials | concept | 25 |
| 271 | AWS SDK configuration: many regions, no leaked keys | concept | 20 |
| 272 | SDK clients: small interfaces and fakes | concept | 30 |
| 273 | SDK clients: spy on StopInstances | concept | 20 |
| 274 | Paginators: walk every page | concept | 25 |
| 275 | Paginators: fix the stuck size report | concept | 20 |
| 276 | AWS API errors: codes, faults and NoSuchKey | concept | 25 |
| 277 | AWS API errors: EC2 dry-run permission checks | concept | 20 |
| 278 | S3 objects: upload, verify and list configs | concept | 30 |
| 279 | S3 objects: fix the log cleanup | concept | 20 |
| 280 | Tag compliance: audit and fix instance tags | concept | 30 |
| 281 | Tag compliance: untagged S3 buckets and bad values | concept | 20 |
| 282 | Drift detection: a tiny terraform plan | concept | 35 |
| 283 | Drift detection: ignore AWS-managed tags | concept | 20 |
| 284 | Practice: cloud inventory report | **practice** · race | 45 |

<!-- end modules -->

## After the course

Some next steps that build on what you've done:

- Turn the *ops inventory CLI* or *uptime monitor* project into a real tool you use at work. Add a `Makefile`, a GitHub Actions workflow and a `goreleaser` config.
- Rebuild the *Kubernetes controller* with [controller-runtime](https://book.kubebuilder.io/) and deploy it to a [kind](https://kind.sigs.k8s.io/) cluster.
- Read [Effective Go](https://go.dev/doc/effective_go) and [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments). Most of it will now feel familiar.
- Practise on [Exercism's Go track](https://exercism.org/tracks/go) to keep your fundamentals sharp.
