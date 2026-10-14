# Groundwork lesson plan

<!-- begin summary -->
398 exercises in 23 modules, about 185 hours of hands-on work.
<!-- end summary -->
The course takes a complete beginner to the Go that DevOps and platform engineers write every day: CLIs, log and config processing, HTTP services, concurrency, Kubernetes controllers and AWS automation. Woven between the Go modules are the tools of the trade, each where it's most useful:

| Track | Comes after | You write |
|---|---|---|
| Linux and the shell | Go foundations (Basics to Testing) | bash scripts and pipelines |
| Git and version control | the shell | git commands against scenario repos |
| Databases and SQL | Text and data | SQL queries, then database/sql in Go |
| Software engineering practice | Databases and SQL | refactorings, reviews and a design doc |
| Containers and CI/CD | Production services | Dockerfiles and CI workflows |
| System design | Containers and CI/CD | load balancers, caches, queues and a design |
| Kubernetes fundamentals | Building services | manifests, kubectl and kustomize |
| Cloud fundamentals | Kubernetes with client-go | IAM policies, network plans and cost estimates |

## How the course is built

Each **concept** has a short run of exercises, so one lesson is never your only exposure to it:

1. **The introduction** (for example *Variables*) teaches the idea from zero, with examples, and keeps its tasks small and guided.
2. **A "fix the bug" round** (*Variables: fix the broken builders*) shows the same idea from another angle, often as broken code or a compiler error to read.
3. **A small realistic job** (*Variables: build ops commands*) needs the idea with less scaffolding. Go modules 1–5 have this third round; later modules usually have two exercises per concept.

Most modules end with a **practice** exercise: one task that combines that module's concepts and adds nothing new. Some end with a **project** instead: a larger program spread over several files (two hours or more), where tests unlock task by task, such as the access log report, the ops inventory CLI or the uptime monitor. The *Capstones* combine the whole course.

**Guides** teach the computer science underneath the exercises: one idea each, from *How a program runs* to *Concurrency* and *Declarative systems and control loops*. Each module lists its guides below. Read a module's guides before or alongside its exercises: press `g` in the watcher, run `groundwork guides`, or open `guides/` in your course folder. Every guide ends with *Check yourself* questions; answer them out loud before reading the answers.

After you pass an exercise, press `s` to compare your code with the reference solution (also saved in `solutions/`). Reading a second way to solve the same problem is one of the fastest ways to learn.

## Suggested schedule

| Pace | Time per week | Finishes in about |
|---|---|---|
| Steady | 5 hours (an hour a weekday) | 6 months |
| Focused | 10 hours | 12 weeks |
| Bootcamp | 20 hours | 6 weeks |

Habits that make it stick (the why is in [LEARNING.md](LEARNING.md)):

- **Read the README first** (`i`). Each one teaches everything its tests check.
- **Use hints in order** (`h`). The first hint is a nudge and the last is close to the answer. Try for 5–10 minutes before each one.
- **Stuck after the Answer hint?** Press `s` for the solution, read it, and type it in yourself. The exercise comes back as a review two days later, so you solve it on your own then. The list marks it ◐.
- **Space the repetitions.** Do a concept's introduction, then come back to its follow-ups the next day, before you've fully forgotten it.
- **Do your reviews.** Practice exercises come back 2, 7 and 30 days after you pass them (`v` in the watcher). Solve them from memory: the effort is what makes it stick.
- **Flashcards daily.** `groundwork cards` turns the guides' questions into an Anki deck; five minutes a day on your phone keeps the ideas fresh.
- **Pair a guide with practice.** Read one section of a guide, then do an exercise. Short reading, then doing, beats a long reading session.
- **Explain it out loud.** After each module, say in two or three sentences what each concept is for. If you can't, revisit its introduction.

<!-- Generated from content/course.json; regenerate with:
     go test ./content -run TestLessonPlan -args -update -->
<!-- begin modules -->
## Phase: Go foundations (about 48 hours)

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

### Module 6: Testing

*About 6.0 hours, in `exercises/06_testing/`.* Now you write the tests. Table tests, subtests, golden files, fakes, benchmarks and fuzzing.

**After this module you can:** Write your own tests: table tests, subtests, golden files, fakes, benchmarks and fuzzing. Your tests are graded by whether they catch real bugs.

**Guides:** [Testing and correctness](content/guides/testing_correctness.md)

| # | Exercise | Kind | Min |
|---:|---|---|---:|
| 129 | Your first test | concept | 20 |
| 130 | Your first test: fix a broken test | concept | 20 |
| 131 | Table-driven tests | concept | 25 |
| 132 | Table-driven tests: error cases | concept | 20 |
| 133 | Subtests and helpers | concept | 25 |
| 134 | Subtests and helpers: files and t.TempDir | concept | 20 |
| 135 | Test fixtures and golden files | concept | 30 |
| 136 | Golden files: write your own | concept | 20 |
| 137 | Fakes and dependency injection | concept | 30 |
| 138 | Fakes: a fake clock for retries | concept | 20 |
| 139 | Benchmarks | concept | 25 |
| 140 | Benchmarks: sub-benchmarks and preallocation | concept | 20 |
| 141 | Fuzzing | concept | 30 |
| 142 | Fuzzing: properties of a truncated alert | concept | 20 |
| 143 | Practice: reproduce a bug report | **practice** | 35 |

## Phase: Everyday tools (about 13 hours)

### Module 7: Linux and the shell

*About 6.5 hours, in `exercises/07_shell/`.* Bash and the classic text tools: the glue of every server, container and CI job.

**After this module you can:** Move files around and redirect output safely, write bash scripts with variables, the environment, quoting, conditions, loops and functions, and slice logs with grep, sed, awk, sort, uniq, cut, find and xargs.

**Guides:** [The operating system for programmers](content/guides/operating_systems.md) · [The shell and the command line](content/guides/shell.md)

| # | Exercise | Kind | Min |
|---:|---|---|---:|
| 144 | Shell commands and quoting | concept | 15 |
| 145 | Files and folders: paths, mkdir, cp, mv and rm | concept | 25 |
| 146 | Shell variables | concept | 25 |
| 147 | Shell variables: fix the quoting bugs | concept | 20 |
| 148 | Exit codes, && and || | concept | 20 |
| 149 | Redirection: >, >>, 2> and here-docs | concept | 25 |
| 150 | if and test | concept | 25 |
| 151 | if and test: fix the backup checks | concept | 15 |
| 152 | Shell loops | concept | 25 |
| 153 | Shell functions | concept | 25 |
| 154 | Grep and pipes | concept | 25 |
| 155 | Sort and uniq: top-N reports | concept | 25 |
| 156 | Columns with cut and awk | concept | 30 |
| 157 | Editing text with sed | concept | 25 |
| 158 | Find and xargs: clean up old logs | concept | 25 |
| 159 | Practice: log triage report | **practice** | 40 |

### Module 8: Git and version control

*About 6.1 hours, in `exercises/08_git/`.* How Git stores history, and the everyday moves: commit, branch, merge, rebase, undo, share and investigate.

**After this module you can:** Commit, stage and ignore files, branch, merge and rebase with confidence, resolve conflicts, undo mistakes safely with revert, reset and reflog, share work through a remote with clone, push, fetch and pull, and find when a bug arrived with log -S, blame and bisect.

**Guides:** [How Git works](content/guides/git.md)

| # | Exercise | Kind | Min |
|---:|---|---|---:|
| 160 | Commits: init, add and commit | concept | 25 |
| 161 | Staging: choose what goes in a commit | concept | 20 |
| 162 | .gitignore: keep build output and secrets out | concept | 25 |
| 163 | Branches: create, switch and merge | concept | 20 |
| 164 | Branches: resolve a merge conflict | concept | 30 |
| 165 | Rebase: catch a branch up with main | concept | 25 |
| 166 | Commits: fix the last commit | concept | 25 |
| 167 | Undoing: revert a bad deploy | concept | 25 |
| 168 | Undoing: rescue lost work with reflog | concept | 25 |
| 169 | Remotes: clone, push and fetch | concept | 30 |
| 170 | Remotes: fix a rejected push | concept | 25 |
| 171 | Investigating history: log -S, blame and diff | concept | 30 |
| 172 | Bisect: find the commit that broke the check | concept | 25 |
| 173 | Practice: release day | **practice** | 35 |

## Phase: Practical Go (about 36 hours)

### Module 9: Text and data

*About 8.5 hours, in `exercises/09_text_and_data/`.* Reading lines, regular expressions, time, JSON, CSV and templates: the daily bread of ops tooling.

**After this module you can:** Parse logs and configs: bufio scanning, regular expressions, time, JSON, CSV and text/template. The project produces an access-log report, checked end to end by your own golden-file test.

**Guides:** [Data formats and parsing](content/guides/data_formats.md)

| # | Exercise | Kind | Min |
|---:|---|---|---:|
| 174 | Reading input line by line: a hosts list | concept | 25 |
| 175 | Reading input line by line: scanning logs | concept | 20 |
| 176 | Regular expressions: release tags and log lines | concept | 30 |
| 177 | Regular expressions: fix the patterns | concept | 20 |
| 178 | Time and durations: cert expiry, uptime, maintenance windows | concept | 30 |
| 179 | Time and durations: fix the backup scheduler | concept | 20 |
| 180 | Encoding JSON: a deployment status report | concept | 30 |
| 181 | Encoding JSON: fix the alert stream | concept | 20 |
| 182 | Decoding JSON: load a service config | concept | 30 |
| 183 | Decoding JSON: fix the kubectl pod reader | concept | 20 |
| 184 | CSV files: a host inventory | concept | 25 |
| 185 | CSV files: fix the TSV DNS export | concept | 20 |
| 186 | Text templates: render an nginx upstream block | concept | 30 |
| 187 | Text templates: systemd unit files | concept | 20 |
| 188 | Practice: .env file parser | **practice** | 35 |
| 189 | Project: access log report | **project** | 135 |

### Module 10: Databases and SQL

*About 6.7 hours, in `exercises/10_sql/`.* Relational databases: tables, queries, joins, indexes and transactions, then SQL from Go.

**After this module you can:** Model data in tables, query it with SELECT, WHERE, JOIN and GROUP BY, use indexes and transactions, and call a database safely from Go with database/sql.

**Guides:** [Relational databases and SQL](content/guides/databases.md)

| # | Exercise | Kind | Min |
|---:|---|---|---:|
| 190 | SELECT and WHERE: find the failed deploys | concept | 25 |
| 191 | SELECT and WHERE: fix the broken filters | concept | 20 |
| 192 | ORDER BY and LIMIT: the latest deploys | concept | 20 |
| 193 | Aggregates and GROUP BY: failures per service | concept | 25 |
| 194 | GROUP BY and HAVING: fix the flaky-service report | concept | 25 |
| 195 | Joins: services, teams and who to page | concept | 30 |
| 196 | Left joins: services with no on-call | concept | 30 |
| 197 | NULL: fix the missing-value bugs | concept | 25 |
| 198 | CREATE TABLE: keys and constraints | concept | 35 |
| 199 | INSERT, UPDATE and DELETE: record a release | concept | 30 |
| 200 | Indexes: make the deploy lookups fast | concept | 30 |
| 201 | database/sql: query deploys from Go | concept | 35 |
| 202 | database/sql: fix the SQL injection | concept | 30 |
| 203 | Practice: incident report queries | **practice** | 40 |

### Module 11: Software engineering practice

*About 5.1 hours, in `exercises/11_craft/`.* Code other people can read and change: naming, simplicity, refactoring, reviews and design.

**After this module you can:** Refactor working code safely under tests, name and structure code for readers, spot problems in review, and write a short design doc.

**Guides:** [Code for humans: naming, simplicity and refactoring](content/guides/code_quality.md) · [Code review and working on a team](content/guides/code_review.md)

| # | Exercise | Kind | Min |
|---:|---|---|---:|
| 204 | Naming things | concept | 25 |
| 205 | Naming things: fix the misleading names | concept | 20 |
| 206 | Extract function | concept | 30 |
| 207 | Guard clauses | concept | 25 |
| 208 | Guard clauses: fix the flipped conditions | concept | 20 |
| 209 | Named constants | concept | 20 |
| 210 | Removing duplication | concept | 25 |
| 211 | Lookup tables | concept | 25 |
| 212 | Code review: catch the bugs in a PR | concept | 35 |
| 213 | Design docs: propose a log retention fix | concept | 40 |
| 214 | Practice: refactor the log rotator | **practice** | 40 |

### Module 12: CLI tools and the OS

*About 8.5 hours, in `exercises/12_cli/`.* Arguments, flags, streams, files, config and archives: build the kind of tool you would put in a pipeline.

**After this module you can:** Build pipeline-ready CLIs: args, flags, subcommands, streams and exit codes, safe file writes, directory walks, config precedence, checksums and archives. The practice combines them in a preflight check tool.

**Guides:** [Security basics for automation](content/guides/security.md)

| # | Exercise | Kind | Min |
|---:|---|---|---:|
| 215 | Command-line arguments: os.Args and usage errors | concept | 20 |
| 216 | Command-line arguments: fix the port waiter | concept | 20 |
| 217 | Flags: parsing options with a FlagSet | concept | 25 |
| 218 | Flags: repeatable and validated flags | concept | 25 |
| 219 | Subcommands: dispatching git-style commands | concept | 30 |
| 220 | Subcommands: a command table and generated help | concept | 25 |
| 221 | Standard streams: stdin, stdout, stderr and exit codes | concept | 25 |
| 222 | Standard streams: grep-style exit codes | concept | 20 |
| 223 | Files: read, scan and write | concept | 25 |
| 224 | Files: an append-only audit log | concept | 20 |
| 225 | Atomic writes: temp file and rename | concept | 30 |
| 226 | Atomic writes: safely change one setting | concept | 25 |
| 227 | Walking directories: find big log files | concept | 25 |
| 228 | Walking directories: fix the backup cleaner | concept | 25 |
| 229 | Config precedence: defaults, file, env and flags | concept | 30 |
| 230 | Config precedence: fix the backup job config | concept | 30 |
| 231 | Checksums and archives: verify and extract safely | concept | 35 |
| 232 | Checksums and archives: publish a release | concept | 30 |
| 233 | Practice: preflight check CLI | **practice** | 45 |

### Module 13: Packages and project layout

*About 6.8 hours, in `exercises/13_packages/`.* Split code into packages, control what is exported, embed files and stamp version info into binaries.

**After this module you can:** Split a program into packages, control your public API, use internal/, embed files and stamp version info into binaries. The project lays out a multi-package ops inventory CLI.

**Guides:** [Modularity and dependencies](content/guides/modularity.md)

| # | Exercise | Kind | Min |
|---:|---|---|---:|
| 234 | Packages and imports | concept | 25 |
| 235 | Packages: two packages, one name | concept | 20 |
| 236 | Exported and unexported names | concept | 20 |
| 237 | Exported names: design a small API | concept | 20 |
| 238 | internal packages | concept | 20 |
| 239 | internal packages: the boundary | concept | 20 |
| 240 | Embedding files | concept | 25 |
| 241 | Embedding files: database migrations | concept | 20 |
| 242 | Version info and ldflags | concept | 25 |
| 243 | Build info: what's inside the binary | concept | 20 |
| 244 | Practice: lay out a tool | **practice** | 40 |
| 245 | Project: ops inventory CLI | **project** | 150 |

## Phase: Building services (about 49 hours)

### Module 14: Networking and HTTP

*About 10.0 hours, in `exercises/14_networking/`.* IP addresses, TCP, DNS, HTTP clients and servers, JSON APIs, middleware and TLS.

**After this module you can:** Work with IPs and CIDRs, TCP, port checks, DNS lookups and their failures, HTTP clients and servers, JSON APIs, middleware and TLS, all tested with httptest.

**Guides:** [Networking from the ground up](content/guides/networking.md)

| # | Exercise | Kind | Min |
|---:|---|---|---:|
| 246 | IP addresses: classify, allowlist, overlap | concept | 25 |
| 247 | CIDRs: normalize and audit subnet rules | concept | 20 |
| 248 | TCP: a line-based control protocol | concept | 30 |
| 249 | TCP: banner grabbing with deadlines | concept | 25 |
| 250 | Port checks: open, closed or filtered | concept | 30 |
| 251 | Port checks: wait for the database | concept | 25 |
| 252 | DNS: lookups, timeouts and failure types | concept | 30 |
| 253 | DNS: TTLs and a migration check | concept | 25 |
| 254 | HTTP clients: fetch a service status | concept | 30 |
| 255 | HTTP clients: fix the deploy notifier | concept | 25 |
| 256 | HTTP servers: handlers and routes | concept | 30 |
| 257 | HTTP servers: fix the log viewer | concept | 30 |
| 258 | Testing HTTP: recorders and fake servers | concept | 25 |
| 259 | Testing HTTP: a regression test for a bug report | concept | 25 |
| 260 | JSON APIs: build a service catalog | concept | 35 |
| 261 | JSON APIs: fix the alerts API | concept | 30 |
| 262 | Middleware: logging, request IDs and panics | concept | 30 |
| 263 | Middleware: chains, status codes and auth | concept | 30 |
| 264 | TLS: trust a CA and read certificates | concept | 30 |
| 265 | TLS: stop skipping verification | concept | 30 |
| 266 | Practice: endpoint health checker | **practice** | 40 |

### Module 15: Concurrency

*About 12.3 hours, in `exercises/15_concurrency/`.* Goroutines, channels, select, sync and context, then subprocesses and signals. Every check runs with the race detector.

**After this module you can:** Use goroutines, channels, select, WaitGroups, mutexes and context; read race reports; run subprocesses with timeouts and stop gracefully on signals; build worker pools and pipelines; collect errors concurrently.

**Guides:** [Concurrency](content/guides/concurrency.md)

| # | Exercise | Kind | Min |
|---:|---|---|---:|
| 267 | Goroutines | concept · race | 20 |
| 268 | Goroutines: fix the maintenance script | concept · race | 15 |
| 269 | Channels | concept · race | 25 |
| 270 | Channels: deadlocks and directions | concept · race | 20 |
| 271 | select and timeouts | concept · race | 25 |
| 272 | select: tickers and deadlines | concept · race | 20 |
| 273 | WaitGroups | concept · race | 20 |
| 274 | WaitGroups: patching the fleet | concept · race | 20 |
| 275 | Mutexes | concept · race | 25 |
| 276 | Mutexes: atomics and slow locks | concept · race | 20 |
| 277 | The race detector | concept · race | 25 |
| 278 | The race detector: hidden races | concept · race | 20 |
| 279 | Context and cancellation | concept · race | 30 |
| 280 | Context: cancelling HTTP requests | concept · race | 25 |
| 281 | Running commands: output, exit codes and timeouts | concept · race | 30 |
| 282 | Running commands: stdin, stderr and PATH lookups | concept · race | 30 |
| 283 | Signals: graceful stop for a rolling drain | concept · race | 30 |
| 284 | Signals: SIGHUP reloads, SIGTERM stops | concept · race | 20 |
| 285 | Worker pools | concept · race | 30 |
| 286 | Worker pools: fix the image puller | concept · race | 20 |
| 287 | Pipelines | concept · race | 30 |
| 288 | Pipelines: fan-in and cancellation | concept · race | 30 |
| 289 | Collecting errors from goroutines | concept · race | 30 |
| 290 | Errors from goroutines: stop on first failure | concept · race | 30 |
| 291 | Project: uptime monitor | **project** · race | 150 |

### Module 16: Production services

*About 11.8 hours, in `exercises/16_production/`.* Logging, metrics, tracing, health checks, graceful shutdown, retries, rate limits and caches: what keeps a service up at 3am.

**After this module you can:** Add what on-call needs: slog logging, Prometheus metrics, trace IDs and alert rules, health checks, graceful shutdown, retries, rate limits, circuit breakers, caches and idempotency.

**Guides:** [Distributed systems basics](content/guides/distributed_systems.md)

| # | Exercise | Kind | Min |
|---:|---|---|---:|
| 292 | Structured logging with slog | concept · race | 25 |
| 293 | Structured logging: access logs | concept · race | 20 |
| 294 | Prometheus-style metrics: counters and /metrics | concept · race | 30 |
| 295 | Prometheus-style metrics: histograms | concept · race | 25 |
| 296 | Observability: trace IDs through logs and calls | concept · race | 30 |
| 297 | Observability: alert rules and an incident walk-through | concept · race | 30 |
| 298 | Health checks: liveness and readiness | concept · race | 25 |
| 299 | Health checks: readiness gates | concept · race | 20 |
| 300 | Graceful shutdown: drain in-flight requests | concept · race | 30 |
| 301 | Graceful shutdown: HTTP server and worker | concept · race | 20 |
| 302 | Retries: backoff and jitter | concept · race | 30 |
| 303 | Retries: retryable HTTP errors | concept · race | 25 |
| 304 | Rate limiting: token buckets | concept · race | 30 |
| 305 | Rate limiting: fix the bucket | concept · race | 25 |
| 306 | Circuit breakers: fail fast | concept · race | 30 |
| 307 | Circuit breakers: protect an HTTP client | concept · race | 25 |
| 308 | TTL caching: expire on read | concept · race | 30 |
| 309 | TTL caching: collapse concurrent loads | concept · race | 25 |
| 310 | Idempotency keys: replay retried requests | concept · race | 30 |
| 311 | Idempotency keys: scope, expire and fingerprint | concept · race | 20 |
| 312 | Project: inventory service | **project** · race | 180 |

### Module 17: Containers and CI/CD

*About 7.2 hours, in `exercises/17_containers/`.* Shipping software: container images, Dockerfiles, registries and pipelines that test and deploy it.

**After this module you can:** Write small, secure multi-stage Dockerfiles, reason about layers, tags and digests, read config from the environment, and define CI pipelines with SHA-pinned actions that test, build, release and promote a Go service that drains cleanly when its container stops.

**Guides:** [Containers and images](content/guides/containers.md) · [Continuous integration and delivery](content/guides/ci_cd.md)

| # | Exercise | Kind | Min |
|---:|---|---|---:|
| 313 | Dockerfiles | concept | 25 |
| 314 | Dockerfiles: fix the slow builds | concept | 25 |
| 315 | Multi-stage builds | concept | 30 |
| 316 | Multi-stage builds: fix the broken images | concept | 25 |
| 317 | Image security: non-root and no secrets | concept | 30 |
| 318 | Image security: fix the audit findings | concept | 25 |
| 319 | Container config: settings from the environment | concept | 35 |
| 320 | CI workflows | concept | 35 |
| 321 | CI workflows: fix the broken pipeline | concept | 25 |
| 322 | CI workflows: a matrix, a cache and artifacts | concept | 30 |
| 323 | Release pipelines | concept | 30 |
| 324 | Release pipelines: fix the risky deploy | concept | 25 |
| 325 | Practice: containerize and ship a service | **practice** · race | 90 |

### Module 18: System design

*About 7.5 hours, in `exercises/18_system_design/`.* How bigger systems fit together: load balancing, caching, queues, sharding and the trade-offs between them.

**After this module you can:** Estimate capacity and set SLOs, then build the core pieces of a bigger system (load balancers, sharding, a consistent-hash ring, an LRU cache, a bounded work queue, an at-least-once consumer) and write down the trade-offs.

**Guides:** [System design fundamentals](content/guides/system_design.md) · [Scaling: caches, queues and load balancers](content/guides/scaling.md)

| # | Exercise | Kind | Min |
|---:|---|---|---:|
| 326 | Capacity estimates: QPS, storage and bandwidth | concept | 30 |
| 327 | SLOs: error budgets and burn rates | concept | 30 |
| 328 | Load balancing: round robin with health checks | concept · race | 30 |
| 329 | Load balancing: fix the least-connections balancer | concept · race | 25 |
| 330 | Sharding: route keys and plan a rebalance | concept | 30 |
| 331 | Hash rings: find a key's node | concept | 30 |
| 332 | Consistent hashing: a ring with virtual nodes | concept | 35 |
| 333 | Consistent hashing: fix the ring | concept | 25 |
| 334 | Linked lists: a deploy queue with container/list | concept | 30 |
| 335 | LRU caches: evict the least recently used | concept · race | 40 |
| 336 | LRU caches: fix the stale cache-aside reads | concept · race | 25 |
| 337 | Work queues: bounded queues and backpressure | concept · race | 30 |
| 338 | Work queues: fix the at-least-once consumer | concept | 30 |
| 339 | Practice: deploy-status service | **practice** · race | 60 |

## Phase: Platform engineering (about 29 hours)

### Module 19: Kubernetes fundamentals

*About 6.0 hours, in `exercises/19_k8s_basics/`.* Kubernetes from the user's side: the objects, kubectl, manifests, rollouts and debugging a broken workload.

**After this module you can:** Write and reason about Pods, Deployments, Services, ConfigMaps, Secrets, probes and resources; read kubectl output to debug a failing rollout; use kustomize; run it all on a local kind cluster.

**Guides:** [Kubernetes from the user's side](content/guides/kubernetes.md) · [Declarative systems and control loops](content/guides/control_loops.md)

| # | Exercise | Kind | Min |
|---:|---|---|---:|
| 340 | Pods: your first manifest | concept | 25 |
| 341 | Pod status: reading kubectl get | concept | 25 |
| 342 | Deployments: replicas, labels and selectors | concept | 25 |
| 343 | Services: selectors, ports and endpoints | concept | 30 |
| 344 | ConfigMaps and Secrets: settings from the cluster | concept | 35 |
| 345 | Probes: readiness, liveness and startup | concept | 25 |
| 346 | Resources: requests, limits and QoS | concept | 30 |
| 347 | Debugging pods: describe, events and exit codes | concept | 30 |
| 348 | Security contexts: least privilege for pods | concept | 25 |
| 349 | Rollouts: surge, unavailable and rollback | concept | 30 |
| 350 | Kustomize: one base, many environments | concept | 35 |
| 351 | Practice: deploy and debug a service | **practice** | 45 |

### Module 20: Kubernetes with client-go

*About 10.9 hours, in `exercises/20_kubernetes/`.* Kubernetes objects, manifests, kubeconfig, clientsets, informers, workqueues and reconcile loops using the real client-go.

**After this module you can:** Build and validate Kubernetes objects, read manifests and kubeconfigs, use the clientset, selectors, watches, informers and workqueues, and write a controller.

| # | Exercise | Kind | Min |
|---:|---|---|---:|
| 352 | Kubernetes objects: building a Deployment | concept | 30 |
| 353 | Kubernetes objects: Services and selectors | concept | 25 |
| 354 | YAML manifests: reading them into Go | concept | 30 |
| 355 | YAML manifests: decoding kinds and linting images | concept | 30 |
| 356 | kubeconfig: clusters, users and contexts | concept | 25 |
| 357 | kubeconfig: building a rest.Config | concept | 20 |
| 358 | The clientset: get, create and delete | concept | 30 |
| 359 | The clientset: updates and create-or-update | concept | 30 |
| 360 | Labels and selectors: matching in Go | concept | 25 |
| 361 | Labels and selectors: selecting on the server | concept | 25 |
| 362 | Watching resources: streaming Pod events | concept · race | 30 |
| 363 | Watching resources: closed watches and ERROR events | concept · race | 25 |
| 364 | Informers: list, watch and cache | concept · race | 35 |
| 365 | Informers: listers and the read-only cache | concept · race | 30 |
| 366 | Workqueues: keys, dedup and the worker loop | concept · race | 30 |
| 367 | Workqueues: retries with backoff | concept · race | 25 |
| 368 | Reconcile loops: desired vs actual | concept | 30 |
| 369 | Reconcile loops: idempotent updates | concept | 30 |
| 370 | Project: Kubernetes controller | **project** · race | 150 |

### Module 21: Cloud fundamentals

*About 5.8 hours, in `exercises/21_cloud/`.* How AWS is put together before you automate it: accounts, regions and zones, IAM, networks, storage, compute, audit and cost.

**After this module you can:** Explain where AWS resources live and who may touch them, write least-privilege IAM policies, plan a VPC's subnets, read CloudTrail events, and estimate and cap cost, all offline.

**Guides:** [The cloud from the ground up](content/guides/cloud.md) · [Cloud and cluster access: IAM, RBAC and short-lived credentials](content/guides/cloud_access.md)

| # | Exercise | Kind | Min |
|---:|---|---|---:|
| 371 | ARNs: accounts, regions and resource scope | concept | 25 |
| 372 | IAM policies: how AWS decides allow or deny | concept | 35 |
| 373 | IAM policies: write least-privilege policies | concept | 30 |
| 374 | IAM policies: fix the over-broad policy | concept | 25 |
| 375 | Assume role: trust policies and temporary credentials | concept | 30 |
| 376 | VPC planning: subnets per availability zone | concept | 30 |
| 377 | Security groups and network ACLs: will this packet get through? | concept | 35 |
| 378 | S3 bucket policies: is this bucket public? | concept | 30 |
| 379 | CloudTrail: trace a leaked access key | concept | 30 |
| 380 | Cloud costs: estimates and budget alarms | concept | 30 |
| 381 | Practice: review an AWS account | **practice** | 45 |

### Module 22: Cloud automation with the AWS SDK

*About 6.4 hours, in `exercises/22_aws/`.* aws-sdk-go-v2 configuration, testable clients, paginators, API errors, S3 and tag compliance, all tested offline.

**After this module you can:** Configure aws-sdk-go-v2 safely, write testable SDK code, paginate, handle API errors, manage S3 objects, audit tags and detect drift.

| # | Exercise | Kind | Min |
|---:|---|---|---:|
| 382 | AWS SDK configuration: regions, profiles and credentials | concept | 25 |
| 383 | AWS SDK configuration: many regions, no leaked keys | concept | 20 |
| 384 | SDK clients: small interfaces and fakes | concept | 30 |
| 385 | SDK clients: spy on StopInstances | concept | 20 |
| 386 | Paginators: walk every page | concept | 25 |
| 387 | Paginators: fix the stuck size report | concept | 20 |
| 388 | AWS API errors: codes, faults and NoSuchKey | concept | 25 |
| 389 | AWS API errors: EC2 dry-run permission checks | concept | 20 |
| 390 | S3 objects: upload, verify and list configs | concept | 30 |
| 391 | S3 objects: fix the log cleanup | concept | 20 |
| 392 | Tag compliance: audit and fix instance tags | concept | 30 |
| 393 | Tag compliance: untagged S3 buckets and bad values | concept | 20 |
| 394 | Drift detection: a tiny terraform plan | concept | 35 |
| 395 | Drift detection: ignore AWS-managed tags | concept | 20 |
| 396 | Practice: cloud inventory report | **practice** · race | 45 |

## Phase: Capstones (about 11 hours)

### Module 23: Capstones

*About 11.0 hours, in `exercises/23_capstones/`.* Two portfolio projects that pull the whole course together: a real service, built, tested, containerised, deployed to Kubernetes and shipped by CI, and the policy checker that guards it.

**After this module you can:** Build a production-style Go service end to end, run it on a local Kubernetes cluster, ship it with CI, enforce deployment policy with your own tool, and present both as portfolio projects.

**Guides:** [Shipping it for real: clusters, registries and your portfolio](content/guides/shipping.md)

| # | Exercise | Kind | Min |
|---:|---|---|---:|
| 397 | Capstone: go links, built and shipped | **project** · race | 360 |
| 398 | Capstone: a policy checker for Kubernetes manifests | **project** | 300 |

<!-- end modules -->

## After the course

Some next steps that build on what you've done:

- Turn the *ops inventory CLI* or *uptime monitor* project into a real tool you use at work. Add a `Makefile`, a GitHub Actions workflow and a `goreleaser` config.
- Rebuild the *Kubernetes controller* with [controller-runtime](https://book.kubebuilder.io/) and deploy it to a [kind](https://kind.sigs.k8s.io/) cluster.
- Read [Effective Go](https://go.dev/doc/effective_go) and [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments). Most of it will now feel familiar.
- Practise on [Exercism's Go track](https://exercism.org/tracks/go) to keep your fundamentals sharp.
