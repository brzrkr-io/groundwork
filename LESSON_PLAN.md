# Groundwork lesson plan

<!-- begin summary -->
362 exercises in 20 modules, about 156 hours of hands-on work.
<!-- end summary -->
The course takes a complete beginner to the Go that DevOps and platform engineers write every day: CLIs, log and config processing, HTTP services, concurrency, Kubernetes controllers and AWS automation. Woven between the Go modules are the tools of the trade, each where it's most useful:

| Track | Comes after | You write |
|---|---|---|
| Linux and the shell | Go foundations | bash scripts and pipelines |
| Git and version control | the shell | git commands against scenario repos |
| Databases and SQL | Text and data | SQL queries, then database/sql in Go |
| Software engineering practice | Testing | refactorings, reviews and a design doc |
| Containers and CI/CD | Packages and project layout | Dockerfiles and CI workflows |
| System design | Production services | load balancers, caches, queues and a design |

## How the course is built

Each **concept** has a short run of exercises, so one lesson is never your only exposure to it:

1. **The introduction** (for example *Variables*) teaches the idea from zero, with examples, and keeps its tasks small and guided.
2. **A "fix the bug" round** (*Variables: fix the broken builders*) shows the same idea from another angle, often as broken code or a compiler error to read.
3. **A small realistic job** (*Variables: build ops commands*) needs the idea with less scaffolding. Go modules 1–5 have this third round; later modules usually have two exercises per concept.

Every module ends with a **practice** exercise that combines its concepts. Each phase ends with a multi-file **project**, where tests unlock task by task.

**Guides** teach the computer science underneath the exercises: one idea each, from *How a program runs* to *Concurrency* and *Declarative systems and control loops*. Each module lists its guides below. Read a module's guides before or alongside its exercises: press `g` in the watcher, run `groundwork guides`, or open `guides/` in your course folder. Every guide ends with *Check yourself* questions; answer them out loud before reading the answers.

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
- **Stuck after the last hint?** Press `s` for the solution, read it, then `groundwork reset` the exercise the next day and solve it without looking. The list marks it ◐ until you do.
- **Space the repetitions.** Do a concept's introduction, then come back to its follow-ups the next day, before you've fully forgotten it.
- **Redo practice exercises cold.** After finishing a module, `groundwork reset <practice>` and solve it again without hints.
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

## Phase: Everyday tools (about 11 hours)

### Module 6: Linux and the shell

*About 5.6 hours, in `exercises/06_shell/`.* Bash and the classic text tools: the glue of every server, container and CI job.

**After this module you can:** Write safe bash scripts with variables, quoting, conditions, loops and functions, and slice logs with grep, sed, awk, sort, uniq, cut, find and xargs.

**Guides:** [The shell and the command line](content/guides/shell.md)

| # | Exercise | Kind | Min |
|---:|---|---|---:|
| 129 | Shell commands and quoting | concept | 15 |
| 130 | Shell variables | concept | 20 |
| 131 | Shell variables: fix the quoting bugs | concept | 20 |
| 132 | Exit codes, && and || | concept | 20 |
| 133 | if and test | concept | 25 |
| 134 | if and test: fix the backup checks | concept | 15 |
| 135 | Shell loops | concept | 25 |
| 136 | Shell functions | concept | 25 |
| 137 | Grep and pipes | concept | 25 |
| 138 | Sort and uniq: top-N reports | concept | 25 |
| 139 | Columns with cut and awk | concept | 30 |
| 140 | Editing text with sed | concept | 25 |
| 141 | Find and xargs: clean up old logs | concept | 25 |
| 142 | Practice: log triage report | **practice** | 40 |

### Module 7: Git and version control

*About 5.2 hours, in `exercises/07_git/`.* How Git stores history, and the everyday moves: commit, branch, merge, rebase, undo and investigate.

**After this module you can:** Commit, stage and ignore files, branch, merge and rebase with confidence, resolve conflicts, undo mistakes safely with revert, reset and reflog, and find when a bug arrived with log -S, blame and bisect.

**Guides:** [How Git works](content/guides/git.md)

| # | Exercise | Kind | Min |
|---:|---|---|---:|
| 143 | Commits: init, add and commit | concept | 25 |
| 144 | Commits: fix the last commit | concept | 25 |
| 145 | Staging: choose what goes in a commit | concept | 20 |
| 146 | .gitignore: keep build output and secrets out | concept | 25 |
| 147 | Branches: create, switch and merge | concept | 20 |
| 148 | Branches: resolve a merge conflict | concept | 30 |
| 149 | Rebase: catch a branch up with main | concept | 25 |
| 150 | Undoing: revert a bad deploy | concept | 25 |
| 151 | Undoing: rescue lost work with reflog | concept | 25 |
| 152 | Investigating history: log -S, blame and diff | concept | 30 |
| 153 | Bisect: find the commit that broke the check | concept | 25 |
| 154 | Practice: release day | **practice** | 35 |

## Phase: Practical Go (about 37 hours)

### Module 8: Text and data

*About 8.2 hours, in `exercises/08_text_and_data/`.* Reading lines, regular expressions, time, JSON, CSV and templates: the daily bread of ops tooling.

**After this module you can:** Parse logs and configs: bufio scanning, regular expressions, time, JSON, CSV and text/template. The project produces an access-log report.

**Guides:** [Data formats and parsing](content/guides/data_formats.md)

| # | Exercise | Kind | Min |
|---:|---|---|---:|
| 155 | Reading input line by line: a hosts list | concept | 25 |
| 156 | Reading input line by line: scanning logs | concept | 20 |
| 157 | Regular expressions: release tags and log lines | concept | 30 |
| 158 | Regular expressions: fix the patterns | concept | 20 |
| 159 | Time and durations: cert expiry, uptime, maintenance windows | concept | 30 |
| 160 | Time and durations: fix the backup scheduler | concept | 20 |
| 161 | Encoding JSON: a deployment status report | concept | 30 |
| 162 | Encoding JSON: fix the alert stream | concept | 20 |
| 163 | Decoding JSON: load a service config | concept | 30 |
| 164 | Decoding JSON: fix the kubectl pod reader | concept | 20 |
| 165 | CSV files: a host inventory | concept | 25 |
| 166 | CSV files: fix the TSV DNS export | concept | 20 |
| 167 | Text templates: render an nginx upstream block | concept | 30 |
| 168 | Text templates: systemd unit files | concept | 20 |
| 169 | Practice: .env file parser | **practice** | 35 |
| 170 | Project: access log report | **project** | 120 |

### Module 9: Databases and SQL

*About 6.7 hours, in `exercises/09_sql/`.* Relational databases: tables, queries, joins, indexes and transactions, then SQL from Go.

**After this module you can:** Model data in tables, query it with SELECT, WHERE, JOIN and GROUP BY, use indexes and transactions, and call a database safely from Go with database/sql.

**Guides:** [Relational databases and SQL](content/guides/databases.md)

| # | Exercise | Kind | Min |
|---:|---|---|---:|
| 171 | SELECT and WHERE: find the failed deploys | concept | 25 |
| 172 | SELECT and WHERE: fix the broken filters | concept | 20 |
| 173 | ORDER BY and LIMIT: the latest deploys | concept | 20 |
| 174 | Aggregates and GROUP BY: failures per service | concept | 25 |
| 175 | GROUP BY and HAVING: fix the flaky-service report | concept | 25 |
| 176 | JOINs: services, teams and who to page | concept | 30 |
| 177 | LEFT JOIN: services with no on-call | concept | 30 |
| 178 | NULL: fix the missing-value bugs | concept | 25 |
| 179 | CREATE TABLE: keys and constraints | concept | 35 |
| 180 | INSERT, UPDATE and DELETE: record a release | concept | 30 |
| 181 | Indexes: make the deploy lookups fast | concept | 30 |
| 182 | database/sql: query deploys from Go | concept | 35 |
| 183 | database/sql: fix the SQL injection | concept | 30 |
| 184 | Practice: incident report queries | **practice** | 40 |

### Module 10: Testing

*About 6.0 hours, in `exercises/10_testing/`.* Now you write the tests. Table tests, subtests, golden files, fakes, benchmarks and fuzzing.

**After this module you can:** Write your own tests: table tests, subtests, golden files, fakes, benchmarks and fuzzing. Your tests are graded by whether they catch real bugs.

**Guides:** [Testing and correctness](content/guides/testing_correctness.md)

| # | Exercise | Kind | Min |
|---:|---|---|---:|
| 185 | Your first test | concept | 20 |
| 186 | Your first test: fix a broken test | concept | 20 |
| 187 | Table-driven tests | concept | 25 |
| 188 | Table-driven tests: error cases | concept | 20 |
| 189 | Subtests and helpers | concept | 25 |
| 190 | Subtests and helpers: files and t.TempDir | concept | 20 |
| 191 | Test fixtures and golden files | concept | 30 |
| 192 | Golden files: write your own | concept | 20 |
| 193 | Fakes and dependency injection | concept | 30 |
| 194 | Fakes: a fake clock for retries | concept | 20 |
| 195 | Benchmarks | concept | 25 |
| 196 | Benchmarks: sub-benchmarks and preallocation | concept | 20 |
| 197 | Fuzzing | concept | 30 |
| 198 | Fuzzing: properties of a truncated alert | concept | 20 |
| 199 | Practice: reproduce a bug report | **practice** | 35 |

### Module 11: Software engineering practice

*About 5.1 hours, in `exercises/11_craft/`.* Code other people can read and change: naming, simplicity, refactoring, reviews and design.

**After this module you can:** Refactor working code safely under tests, name and structure code for readers, spot problems in review, and write a short design doc.

**Guides:** [Code for humans: naming, simplicity and refactoring](content/guides/code_quality.md) · [Code review and working on a team](content/guides/code_review.md)

| # | Exercise | Kind | Min |
|---:|---|---|---:|
| 200 | Naming things | concept | 25 |
| 201 | Naming things: fix the misleading names | concept | 20 |
| 202 | Extract function | concept | 30 |
| 203 | Guard clauses | concept | 25 |
| 204 | Guard clauses: fix the flipped conditions | concept | 20 |
| 205 | Named constants | concept | 20 |
| 206 | Removing duplication | concept | 25 |
| 207 | Lookup tables | concept | 25 |
| 208 | Code review: catch the bugs in a PR | concept | 35 |
| 209 | Design docs: propose a log retention fix | concept | 40 |
| 210 | Practice: refactor the log rotator | **practice** | 40 |

### Module 12: Packages and project layout

*About 4.2 hours, in `exercises/12_packages/`.* Split code into packages, control what is exported, embed files and stamp version info into binaries.

**After this module you can:** Split a program into packages, control your public API, use internal/, embed files and stamp version info into binaries.

**Guides:** [Modularity and dependencies](content/guides/modularity.md)

| # | Exercise | Kind | Min |
|---:|---|---|---:|
| 211 | Packages and imports | concept | 25 |
| 212 | Packages: two packages, one name | concept | 20 |
| 213 | Exported and unexported names | concept | 20 |
| 214 | Exported names: design a small API | concept | 20 |
| 215 | internal packages | concept | 20 |
| 216 | internal packages: the boundary | concept | 20 |
| 217 | Embedding files | concept | 25 |
| 218 | Embedding files: database migrations | concept | 20 |
| 219 | Version info and ldflags | concept | 25 |
| 220 | Build info: what's inside the binary | concept | 20 |
| 221 | Practice: lay out a tool | **practice** | 40 |

### Module 13: Containers and CI/CD

*About 6.3 hours, in `exercises/13_containers/`.* Shipping software: container images, Dockerfiles, registries and pipelines that test and deploy it.

**After this module you can:** Write small, secure multi-stage Dockerfiles, reason about layers, tags and digests, read config from the environment, and define CI pipelines that test, build, release and promote a Go service.

**Guides:** [Containers and images](content/guides/containers.md) · [Continuous integration and delivery](content/guides/ci_cd.md)

| # | Exercise | Kind | Min |
|---:|---|---|---:|
| 222 | Dockerfiles | concept | 25 |
| 223 | Dockerfiles: fix the slow builds | concept | 25 |
| 224 | Multi-stage builds | concept | 30 |
| 225 | Multi-stage builds: fix the broken images | concept | 25 |
| 226 | Image security: non-root and no secrets | concept | 30 |
| 227 | Image security: fix the audit findings | concept | 25 |
| 228 | Container config: settings from the environment | concept | 35 |
| 229 | CI workflows | concept | 30 |
| 230 | CI workflows: fix the broken pipeline | concept | 25 |
| 231 | CI workflows: a matrix, a cache and artifacts | concept | 30 |
| 232 | Release pipelines | concept | 30 |
| 233 | Release pipelines: fix the risky deploy | concept | 25 |
| 234 | Practice: containerize and ship a service | **practice** | 45 |

## Phase: Building tools and services (about 49 hours)

### Module 14: CLI tools and the OS

*About 12.1 hours, in `exercises/14_cli/`.* Arguments, flags, files, environment, subprocesses and signals: build the kind of tool you would put in a pipeline.

**After this module you can:** Build pipeline-ready CLIs: args, flags, subcommands, streams and exit codes, safe file writes, directory walks, config precedence, subprocesses, signals, checksums and archives.

**Guides:** [The operating system for programmers](content/guides/operating_systems.md)

| # | Exercise | Kind | Min |
|---:|---|---|---:|
| 235 | Command-line arguments: os.Args and usage errors | concept | 20 |
| 236 | Command-line arguments: fix the port waiter | concept | 20 |
| 237 | Flags: parsing options with a FlagSet | concept | 25 |
| 238 | Flags: repeatable and validated flags | concept | 25 |
| 239 | Subcommands: dispatching git-style commands | concept | 30 |
| 240 | Subcommands: a command table and generated help | concept | 25 |
| 241 | Standard streams: stdin, stdout, stderr and exit codes | concept | 25 |
| 242 | Standard streams: grep-style exit codes | concept | 20 |
| 243 | Files: read, scan and write | concept | 25 |
| 244 | Files: an append-only audit log | concept | 20 |
| 245 | Atomic writes: temp file and rename | concept | 30 |
| 246 | Atomic writes: safely change one setting | concept | 25 |
| 247 | Walking directories: find big log files | concept | 25 |
| 248 | Walking directories: fix the backup cleaner | concept | 25 |
| 249 | Config precedence: defaults, file, env and flags | concept | 30 |
| 250 | Config precedence: fix the backup job config | concept | 30 |
| 251 | Running commands: output, exit codes and timeouts | concept | 30 |
| 252 | Running commands: stdin, stderr and PATH lookups | concept | 30 |
| 253 | Signals: graceful stop for a rolling drain | concept | 30 |
| 254 | Signals: SIGHUP reloads, SIGTERM stops | concept | 20 |
| 255 | Checksums and archives: verify and extract safely | concept | 35 |
| 256 | Checksums and archives: publish a release | concept | 30 |
| 257 | Project: ops inventory CLI | **project** | 150 |

### Module 15: Networking and HTTP

*About 9.1 hours, in `exercises/15_networking/`.* IP addresses, TCP, HTTP clients and servers, JSON APIs, middleware and TLS.

**After this module you can:** Work with IPs and CIDRs, TCP, port checks, HTTP clients and servers, JSON APIs, middleware and TLS, all tested with httptest.

**Guides:** [Networking from the ground up](content/guides/networking.md)

| # | Exercise | Kind | Min |
|---:|---|---|---:|
| 258 | IP addresses: classify, allowlist, overlap | concept | 25 |
| 259 | CIDRs: normalize and audit subnet rules | concept | 20 |
| 260 | TCP: a line-based control protocol | concept | 30 |
| 261 | TCP: banner grabbing with deadlines | concept | 25 |
| 262 | Port checks: open, closed or filtered | concept | 30 |
| 263 | Port checks: wait for the database | concept | 25 |
| 264 | HTTP clients: fetch a service status | concept | 30 |
| 265 | HTTP clients: fix the deploy notifier | concept | 25 |
| 266 | HTTP servers: handlers and routes | concept | 30 |
| 267 | HTTP servers: fix the log viewer | concept | 30 |
| 268 | Testing HTTP: recorders and fake servers | concept | 25 |
| 269 | Testing HTTP: a regression test for a bug report | concept | 25 |
| 270 | JSON APIs: build a service catalog | concept | 35 |
| 271 | JSON APIs: fix the alerts API | concept | 30 |
| 272 | Middleware: logging, request IDs and panics | concept | 30 |
| 273 | Middleware: chains, status codes and auth | concept | 30 |
| 274 | TLS: trust a CA and read certificates | concept | 30 |
| 275 | TLS: stop skipping verification | concept | 30 |
| 276 | Practice: endpoint health checker | **practice** | 40 |

### Module 16: Concurrency

*About 10.5 hours, in `exercises/16_concurrency/`.* Goroutines, channels, select, sync and context. Every check runs with the race detector.

**After this module you can:** Use goroutines, channels, select, WaitGroups, mutexes and context; read race reports; build worker pools and pipelines; collect errors concurrently.

**Guides:** [Concurrency](content/guides/concurrency.md)

| # | Exercise | Kind | Min |
|---:|---|---|---:|
| 277 | Goroutines | concept · race | 20 |
| 278 | Goroutines: fix the maintenance script | concept · race | 15 |
| 279 | Channels | concept · race | 25 |
| 280 | Channels: deadlocks and directions | concept · race | 20 |
| 281 | select and timeouts | concept · race | 25 |
| 282 | select: tickers and deadlines | concept · race | 20 |
| 283 | WaitGroups | concept · race | 20 |
| 284 | WaitGroups: patching the fleet | concept · race | 20 |
| 285 | Mutexes | concept · race | 25 |
| 286 | Mutexes: atomics and slow locks | concept · race | 20 |
| 287 | The race detector | concept · race | 25 |
| 288 | The race detector: hidden races | concept · race | 20 |
| 289 | Context and cancellation | concept · race | 30 |
| 290 | Context: cancelling HTTP requests | concept · race | 25 |
| 291 | Worker pools | concept · race | 30 |
| 292 | Worker pools: fix the image puller | concept · race | 20 |
| 293 | Pipelines | concept · race | 30 |
| 294 | Pipelines: fan-in and cancellation | concept · race | 30 |
| 295 | Collecting errors from goroutines | concept · race | 30 |
| 296 | Errors from goroutines: stop on first failure | concept · race | 30 |
| 297 | Project: uptime monitor | **project** · race | 150 |

### Module 17: Production services

*About 10.8 hours, in `exercises/17_production/`.* Logging, metrics, health checks, graceful shutdown, retries, rate limits and caches: what keeps a service up at 3am.

**After this module you can:** Add what on-call needs: slog logging, Prometheus metrics, health checks, graceful shutdown, retries, rate limits, circuit breakers, caches and idempotency.

**Guides:** [Distributed systems basics](content/guides/distributed_systems.md)

| # | Exercise | Kind | Min |
|---:|---|---|---:|
| 298 | Structured logging with slog | concept · race | 25 |
| 299 | Structured logging: access logs | concept · race | 20 |
| 300 | Prometheus-style metrics: counters and /metrics | concept · race | 30 |
| 301 | Prometheus-style metrics: histograms | concept · race | 25 |
| 302 | Health checks: liveness and readiness | concept · race | 25 |
| 303 | Health checks: readiness gates | concept · race | 20 |
| 304 | Graceful shutdown: drain in-flight requests | concept · race | 30 |
| 305 | Graceful shutdown: HTTP server and worker | concept · race | 20 |
| 306 | Retries: backoff and jitter | concept · race | 30 |
| 307 | Retries: retryable HTTP errors | concept · race | 25 |
| 308 | Rate limiting: token buckets | concept · race | 30 |
| 309 | Rate limiting: fix the bucket | concept · race | 25 |
| 310 | Circuit breakers: fail fast | concept · race | 30 |
| 311 | Circuit breakers: protect an HTTP client | concept · race | 25 |
| 312 | TTL caching: expire on read | concept · race | 30 |
| 313 | TTL caching: collapse concurrent loads | concept · race | 25 |
| 314 | Idempotency keys: replay retried requests | concept · race | 30 |
| 315 | Idempotency keys: scope, expire and fingerprint | concept · race | 20 |
| 316 | Project: inventory service | **project** · race | 180 |

### Module 18: System design

*About 6.4 hours, in `exercises/18_system_design/`.* How bigger systems fit together: load balancing, caching, queues, sharding and the trade-offs between them.

**After this module you can:** Estimate capacity and set SLOs, then build the core pieces of a bigger system (load balancers, sharding, a consistent-hash ring, an LRU cache, a bounded work queue, an at-least-once consumer) and write down the trade-offs.

**Guides:** [System design fundamentals](content/guides/system_design.md) · [Scaling: caches, queues and load balancers](content/guides/scaling.md)

| # | Exercise | Kind | Min |
|---:|---|---|---:|
| 317 | Capacity estimates: QPS, storage and bandwidth | concept | 30 |
| 318 | SLOs: error budgets and burn rates | concept | 30 |
| 319 | Load balancing: round robin with health checks | concept · race | 30 |
| 320 | Load balancing: fix the least-connections balancer | concept · race | 25 |
| 321 | Sharding: route keys and plan a rebalance | concept | 30 |
| 322 | Consistent hashing: a ring with virtual nodes | concept | 35 |
| 323 | Consistent hashing: fix the ring | concept | 25 |
| 324 | LRU caches: evict the least recently used | concept · race | 35 |
| 325 | LRU caches: fix the stale cache-aside reads | concept · race | 25 |
| 326 | Work queues: bounded queues and backpressure | concept · race | 30 |
| 327 | Work queues: fix the at-least-once consumer | concept | 30 |
| 328 | Practice: deploy-status service | **practice** · race | 60 |

## Phase: Platform engineering (about 17 hours)

### Module 19: Kubernetes with client-go

*About 10.9 hours, in `exercises/19_kubernetes/`.* Kubernetes objects, manifests, kubeconfig, clientsets, informers, workqueues and reconcile loops using the real client-go.

**After this module you can:** Build and validate Kubernetes objects, read manifests and kubeconfigs, use the clientset, selectors, watches, informers and workqueues, and write a controller.

**Guides:** [Declarative systems and control loops](content/guides/control_loops.md)

| # | Exercise | Kind | Min |
|---:|---|---|---:|
| 329 | Kubernetes objects: building a Deployment | concept | 30 |
| 330 | Kubernetes objects: Services and selectors | concept | 25 |
| 331 | YAML manifests: reading them into Go | concept | 30 |
| 332 | YAML manifests: decoding kinds and linting images | concept | 30 |
| 333 | kubeconfig: clusters, users and contexts | concept | 25 |
| 334 | kubeconfig: building a rest.Config | concept | 20 |
| 335 | The clientset: get, create and delete | concept | 30 |
| 336 | The clientset: updates and create-or-update | concept | 30 |
| 337 | Labels and selectors: matching in Go | concept | 25 |
| 338 | Labels and selectors: selecting on the server | concept | 25 |
| 339 | Watching resources: streaming Pod events | concept · race | 30 |
| 340 | Watching resources: closed watches and ERROR events | concept · race | 25 |
| 341 | Informers: list, watch and cache | concept · race | 35 |
| 342 | Informers: listers and the read-only cache | concept · race | 30 |
| 343 | Workqueues: keys, dedup and the worker loop | concept · race | 30 |
| 344 | Workqueues: retries with backoff | concept · race | 25 |
| 345 | Reconcile loops: desired vs actual | concept | 30 |
| 346 | Reconcile loops: idempotent updates | concept | 30 |
| 347 | Project: Kubernetes controller | **project** · race | 150 |

### Module 20: Cloud automation with the AWS SDK

*About 6.4 hours, in `exercises/20_aws/`.* aws-sdk-go-v2 configuration, testable clients, paginators, API errors, S3 and tag compliance, all tested offline.

**After this module you can:** Configure aws-sdk-go-v2 safely, write testable SDK code, paginate, handle API errors, manage S3 objects, audit tags and detect drift.

**Guides:** [Security basics for automation](content/guides/security.md)

| # | Exercise | Kind | Min |
|---:|---|---|---:|
| 348 | AWS SDK configuration: regions, profiles and credentials | concept | 25 |
| 349 | AWS SDK configuration: many regions, no leaked keys | concept | 20 |
| 350 | SDK clients: small interfaces and fakes | concept | 30 |
| 351 | SDK clients: spy on StopInstances | concept | 20 |
| 352 | Paginators: walk every page | concept | 25 |
| 353 | Paginators: fix the stuck size report | concept | 20 |
| 354 | AWS API errors: codes, faults and NoSuchKey | concept | 25 |
| 355 | AWS API errors: EC2 dry-run permission checks | concept | 20 |
| 356 | S3 objects: upload, verify and list configs | concept | 30 |
| 357 | S3 objects: fix the log cleanup | concept | 20 |
| 358 | Tag compliance: audit and fix instance tags | concept | 30 |
| 359 | Tag compliance: untagged S3 buckets and bad values | concept | 20 |
| 360 | Drift detection: a tiny terraform plan | concept | 35 |
| 361 | Drift detection: ignore AWS-managed tags | concept | 20 |
| 362 | Practice: cloud inventory report | **practice** · race | 45 |

<!-- end modules -->

## After the course

Some next steps that build on what you've done:

- Turn the *ops inventory CLI* or *uptime monitor* project into a real tool you use at work. Add a `Makefile`, a GitHub Actions workflow and a `goreleaser` config.
- Rebuild the *Kubernetes controller* with [controller-runtime](https://book.kubebuilder.io/) and deploy it to a [kind](https://kind.sigs.k8s.io/) cluster.
- Read [Effective Go](https://go.dev/doc/effective_go) and [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments). Most of it will now feel familiar.
- Practise on [Exercism's Go track](https://exercism.org/tracks/go) to keep your fundamentals sharp.
