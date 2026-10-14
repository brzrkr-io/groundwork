# The operating system for programmers

**In one sentence:** the operating system's *kernel* is the program that owns the machine: it shares the CPU, memory and devices between all other programs, enforces who may do what, and gives your program a small set of doors to the outside world: arguments, environment, files, streams, exit codes and signals.

**Why it matters:** every CLI you write and every container you deploy talks to the OS through exactly those doors. "Permission denied", "exit code 137", "my logs broke the JSON output" and "the pod took 30 seconds to stop" are all OS questions. Knowing the doors turns them from mysteries into five-minute fixes.

**You'll learn:**
- what the kernel does and how programs ask it for things (system calls)
- what a process is, and how parents, children and PIDs fit together
- how files, paths and permissions like `0644` and `0755` work
- stdin, stdout and stderr, redirection and pipes
- arguments, environment variables and exit codes, and why 0 means success
- signals, and what Kubernetes does when it stops your pod

## What the kernel does

Picture one Linux server. nginx, your Go service, a metrics exporter, `sshd` and your shell all run "at the same time" on 4 CPU cores. They all want memory, disk and the network. Who decides?

The *kernel*: the core of the operating system, always running, with full control of the hardware. Everything else runs in *user space*, with no direct access to the hardware at all.

| Resource | The kernel's job | Where you see it |
|---|---|---|
| CPU | switch between programs many times a second | CPU % in `top` |
| memory | give each process its own private memory | the OOM killer |
| devices | hide disks and network cards behind one API | files and sockets |
| permissions | check every request against who is asking | "permission denied" |

Each process gets its own memory that no other process can read. When memory runs out, the kernel's *OOM killer* (out of memory) picks a process and kills it.

**Containers are not small virtual machines.** A container is ordinary processes with kernel-enforced limits: *namespaces* control what they can see (their own PIDs, their own filesystem), and *cgroups* control how much they can use (CPU, memory). A Kubernetes memory limit is a cgroup limit, and hitting it triggers the OOM killer.

## System calls: the kernel's API

Your program can't touch the disk or the network itself. It asks the kernel, through a *system call* (syscall): a request like "open this file", "read 4096 bytes", "start this program", "exit with code 1".

```text
  your Go code        data, err := os.ReadFile("app.yaml")
       │
  Go's os package     openat ─► fstat ─► read ─► close    (system calls)
       │
  ═════╪═════════════════════ kernel boundary ═════════════════════
       ▼
  kernel              checks permissions, finds the file, asks the disk
```

The `strace` tool on Linux shows every syscall a program makes. Here is `cat app.yaml`, with library loading trimmed:

```text
openat(AT_FDCWD, "app.yaml", O_RDONLY)  = 3
read(3, "replicas: 3\n", 262144)        = 12
write(1, "replicas: 3\n", 12)           = 12
read(3, "", 262144)                     = 0
close(3)                                = 0
```

1. `openat` opens the file. The kernel answers `3`: a ticket number for this open file.
2. `read(3, ...)` reads up to 262144 bytes and gets 12.
3. `write(1, ...)` writes those 12 bytes to number 1, which is standard output.
4. The next `read` returns 0 bytes: end of file.
5. `close(3)` gives the ticket back.

Those ticket numbers are file descriptors, coming up below. When a tool hangs or fails with a vague error, `strace` shows you exactly which request the kernel refused.

## Processes

A *process* is a running program: its code, its private memory, its open files and an ID. Run the same binary twice and you get two processes.

Each process has a *PID* (process ID) and a *parent*: the process that started it. When you type `./deployctl` in bash, bash is the parent and `deployctl` is the *child*. On a typical server the tree looks like this:

```text
  PID 1  systemd                started by the kernel at boot
   ├── sshd
   │    └── bash                your login shell
   │         └── deployctl      the program you just ran
   └── containerd-shim
        └── checkout-api        a container's main process
```

`ps -ef` lists processes with their PID and parent PID (PPID). `kill <pid>` sends a signal to one, as you'll see in *Signals* below.

When a child exits, it hands an exit code to its parent. That's how your shell knows whether the command worked.

Inside a container, your program is usually PID 1 of its own namespace. That matters because Kubernetes sends its stop signal to PID 1, and only PID 1.

### Threads, briefly

A process can run several *threads*: separate lines of execution that share the same memory. The kernel schedules threads onto CPU cores. Go spreads your program's work across a few threads for you, which is how one Go process can use every core. The *Concurrency* guide explains why shared memory makes this tricky.

## Files and the filesystem

Unix has a famous design rule: **everything is a file**. Regular files, directories, devices (`/dev/null`, a black hole that swallows writes), live process information (`/proc/<pid>/`), and even network connections are all opened, read and written with the same syscalls.

### Paths

- An *absolute path* starts at the root: `/var/log/nginx/access.log`.
- A *relative path* starts from the process's *working directory*: `logs/app.log` means "`logs/app.log` inside wherever I was started".

A classic ops bug: a script works when you run it by hand, then fails under cron or systemd. They start it in a different working directory, so its relative paths point somewhere else. Use absolute paths in jobs.

### Permissions

`ls -l` shows who may do what:

```text
-rw-r--r--  1  deploy  ops  2048  Oct 13 09:15  app.yaml
│└┬┘└┬┘└┬┘     └──┬─┘  └┬┘
│ │  │  │         │     └── group that owns it
│ │  │  │         └──────── user that owns it
│ │  │  └── others:  r--   read only
│ │  └───── group:   r--   read only
│ └──────── owner:   rw-   read and write
└────────── type: - file, d directory, l link
```

Each group of three is `r` (read), `w` (write), `x` (execute). For a directory, `r` means "may list it" and `x` means "may enter it".

Each letter is worth a number, and you add them up per group: `r` = 4, `w` = 2, `x` = 1. Three groups make three digits, written in *octal* (base 8) with a leading 0:

| Octal | Letters | Typical use |
|---|---|---|
| `0644` | `rw-r--r--` | config files, logs |
| `0755` | `rwxr-xr-x` | programs, directories |
| `0600` | `rw-------` | secrets, private keys, kubeconfigs |
| `0700` | `rwx------` | private directories like `~/.ssh` |

So `0644` is owner 6 (4+2, read and write), group 4 (read), others 4 (read).

Go uses the same numbers. `os.WriteFile("token", data, 0600)` creates a file only its owner can read (the mode applies only when the file is created). The leading `0` makes Go read the number as octal; Go also accepts `0o600`, which means the same and is easier to spot. The process's *umask* can remove bits from what you ask for (a common umask of `022` turns `0666` into `0644`), but it never adds any.

The user `root` bypasses most permission checks. That's why tools shouldn't run as root unless they must; the *Security basics for automation* guide goes further.

### Replacing a file safely

Picture your tool rewriting `nginx.conf` while nginx reloads. `os.WriteFile` first empties the file, then writes. A reader in between sees a half-written config, and a crash leaves it that way.

The fix uses one kernel guarantee: renaming a file over another is *atomic*, meaning all or nothing. Every reader sees the whole old file or the whole new one, never a mix.

```text
  1. write the new content to  nginx.conf.tmp-8812   (same directory)
  2. Sync (wait until it's really on disk), then Close
  3. rename nginx.conf.tmp-8812 → nginx.conf          (one atomic step)
```

Keep the temp file in the **same directory**: rename is only atomic within one filesystem, and `/tmp` is often a different one. You'll build this in *Atomic writes: temp file and rename*.

## File descriptors and the standard streams

A *file descriptor* (fd) is the small number the kernel hands your process for each open file, like the `3` in the `strace` output. Every process starts with three already open:

| fd | Name | Go | Meant for |
|---|---|---|---|
| 0 | stdin | `os.Stdin` | input |
| 1 | stdout | `os.Stdout` | the result: the data |
| 2 | stderr | `os.Stderr` | messages for humans: progress, warnings, errors |

```text
                     ┌─────────────────┐
  keyboard / file ──►│ 0 stdin         │
  / pipe             │                 │
                     │     process     │ 1 stdout ──► terminal / file / pipe
                     │                 │
                     │                 │ 2 stderr ──► terminal
                     └─────────────────┘
```

**The rule: results go to stdout, everything else to stderr.** Then `mytool > hosts.json` gives you clean JSON in the file, while progress messages still reach your screen.

### Redirection

The shell can point those fds somewhere else before the program starts. The program doesn't know or care.

| You type | Effect |
|---|---|
| `cmd > out.txt` | stdout into a file (overwrites it) |
| `cmd >> out.txt` | stdout appended to a file |
| `cmd 2> err.txt` | stderr into a file |
| `cmd > all.log 2>&1` | stdout into the file, then stderr to the same place |
| `cmd < hosts.txt` | stdin read from a file |

`2>&1` means "make fd 2 point wherever fd 1 points **right now**". Order matters: `cmd 2>&1 > all.log` sends stderr to the terminal, because at that moment fd 1 still pointed there.

### Pipes

A *pipe* (`|`) connects one program's stdout to the next program's stdin. To count failed deploy requests:

```bash
cat access.log | grep '"POST /deploy' | grep -c '" 502 '
```

```text
  cat access.log ──stdout──► grep POST ──stdout──► grep -c 502 ──► terminal
                    pipe                  pipe
  every program's stderr still goes straight to the terminal
```

All three programs run at the same time, passing lines along as they go. That's streaming, from the *Data formats and parsing* guide: the 40 GB log never sits in memory.

## Arguments, environment and exit codes

Here is a small CLI that uses every door from this section:

```go
package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: deployctl <service>")
		os.Exit(2)
	}
	service := os.Args[1]

	env := os.Getenv("DEPLOY_ENV")
	if env == "" {
		env = "staging"
	}

	fmt.Fprintf(os.Stderr, "deploying %s to %s...\n", service, env)
	fmt.Printf("%s %s ok\n", service, env)
}
```

Build it with `go build -o deployctl .` and try it:

```bash
$ ./deployctl
usage: deployctl <service>
$ echo $?
2
$ ./deployctl checkout
deploying checkout to staging...
checkout staging ok
$ DEPLOY_ENV=production ./deployctl checkout 2>/dev/null
checkout production ok
$ ./deployctl checkout > result.txt
deploying checkout to staging...
```

1. No argument: the usage message goes to stderr, and the program exits with code 2. `$?` is the shell's variable for the last exit code.
2. One argument, no `DEPLOY_ENV`: it falls back to `staging`. Both lines appear, because stdout and stderr both go to the terminal.
3. `DEPLOY_ENV=production` sets a variable for this one command. `2>/dev/null` throws stderr away, leaving only the result.
4. `> result.txt` sends stdout to the file. Only the progress message stays on screen, and `result.txt` holds `checkout staging ok`.

Under `strace`, the last two lines are `write(2, "deploying checkout to staging..."..., 33)` and `write(1, "checkout staging ok\n", 20)`. `fmt.Fprintf(os.Stderr, ...)` is just a write to fd 2.

### Command-line arguments

`os.Args` is a slice of strings. `os.Args[0]` is the program's own name; the real arguments start at `os.Args[1]`.

The shell splits what you type on spaces and **removes the quotes**. `./deployctl "checkout api"` passes one argument, `checkout api`, with no quote characters in it. `./deployctl checkout api` passes two. Quotes are instructions to the shell, just like quotes in Go source mark where a string starts and ends.

You'll practise this in *Command-line arguments*, and grow it into flags and subcommands.

### Environment variables

An *environment variable* is a `NAME=value` string the process receives at start. Each child gets a **copy** of its parent's environment (in a shell, only the variables you `export`). That's why a script that runs `export REGION=eu-west-1` can't change your shell's `REGION`: it changed its own copy, which vanishes when it exits.

In Kubernetes, the `env:` list in a pod spec becomes exactly these variables. Settings that change per deploy (the database host, the log level) usually travel this way, while secrets deserve more care.

### Exit codes

When a process ends, it hands its parent one number from 0 to 255: the *exit code* (or exit status).

**0 means success.** There is one way to succeed and many ways to fail, so all the other numbers are free to say *what* failed.

| Code | Usual meaning |
|---|---|
| 0 | success |
| 1 | general failure |
| 2 | wrong usage: bad arguments or flags |
| 126 | found, but not executable (check `chmod +x`) |
| 127 | command not found (check `PATH` or a typo) |
| 128 + n | killed by signal n |
| 137 | 128 + 9: killed by SIGKILL (often the OOM killer) |
| 143 | 128 + 15: stopped by SIGTERM |

Only 0 and "not 0" are universal; the rest are conventions. Every tool in the chain trusts them:

- the shell: `make && ./deploy.sh` runs the deploy only if `make` exited 0
- CI: a step that exits non-zero fails the pipeline
- Kubernetes: a container's exit code decides whether a Job counts as succeeded

So **a tool that fails but exits 0 is lying**, and the pipeline goes green on a broken deploy.

Two Go details: `os.Exit` ends the program immediately and skips any deferred functions, so call it only at the end of `main`. And `go run` prints `exit status 2` and exits with 1 itself, so to see your real exit code, build the binary first.

## Signals

A *signal* is a small numbered notice the kernel delivers to a process, interrupting whatever it was doing. Pressing Ctrl-C sends one; so does `kill`.

| Signal | Number | Sent by | Can the program catch it? |
|---|---|---|---|
| SIGINT | 2 | Ctrl-C in a terminal | yes |
| SIGTERM | 15 | `kill`, systemd, Kubernetes | yes: "please stop" |
| SIGKILL | 9 | `kill -9`, the OOM killer | **no**: the kernel ends it |
| SIGHUP | 1 | terminal closed; by convention "reload config" | yes |

A program can *catch* a signal it's allowed to: run its own code instead of dying at once. Catching SIGTERM is how a service shuts down cleanly: stop taking new work, finish what's in flight, flush logs, exit 0.

```go
package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)

	fmt.Println("worker running as PID", os.Getpid())
	sig := <-sigs
	fmt.Println("got signal:", sig)
	fmt.Println("finishing the current job, then exiting")
}
```

1. `make(chan os.Signal, 1)` creates a channel: a pipe that Go values arrive on. `signal.Notify` asks Go to deliver SIGINT and SIGTERM there instead of ending the program.
2. `<-sigs` waits until a value (a signal) arrives.
3. `kill -TERM <pid>` prints `got signal: terminated`, then the last line, and the exit code is 0. Ctrl-C prints `got signal: interrupt`.
4. `kill -KILL <pid>` prints nothing more. The kernel removes the process, and the shell reports `Killed` with exit code 137.

Without `signal.Notify`, a Go program simply dies on SIGTERM, mid-job.

### How Kubernetes stops a pod

```text
  kubectl delete pod, a rolling update, or a node drain
    │
    ├─► pod marked Terminating. The grace period clock starts:
    │   terminationGracePeriodSeconds (default 30)
    ├─► at the same time, it's removed from its Services,
    │   so new traffic stops (this can take a few seconds)
    ├─► preStop hook runs, if the pod defines one
    ├─► SIGTERM to PID 1 in each container
    └─► grace period over and still running? SIGKILL. Exit code 137.
```

If your program ignores SIGTERM, every deploy waits the full 30 seconds per pod, then kills it mid-request. The same happens when a shell is PID 1 (a Dockerfile's shell-form `CMD`): it doesn't pass SIGTERM on to your program, as the *Containers and images* guide explains. You'll handle signals properly in *Signals: graceful stop for a rolling drain* and later in *Graceful shutdown: drain in-flight requests*.

## Common mistakes

- **Logging to stdout.** Progress messages end up inside your JSON output. Send them to stderr.
- **Writing `2>&1` before `>`.** `cmd 2>&1 > file` leaves stderr on the terminal. Write `cmd > file 2>&1`.
- **Exiting 0 after a failure.** CI and `&&` believe you. Exit non-zero on any error.
- **Calling `os.Exit` deep inside the program.** It skips deferred cleanup. Return errors up to `main` and exit there.
- **Relative paths in cron jobs or systemd units.** The working directory isn't what you expect. Use absolute paths.
- **`chmod 777` to "fix" permission denied.** Now everyone can change your file. Find out which user needs which permission.
- **Secrets readable by everyone.** Use `0600` for keys, tokens and kubeconfigs.
- **Ignoring SIGTERM.** Kubernetes waits out the grace period, then SIGKILLs you mid-request.

## Check yourself

1. Your tool prints a JSON report and also "fetching 40 hosts...". A teammate runs `mytool > report.json` and the JSON is broken. What's wrong, and what's the fix?
2. What does the file permission `0640` allow, and for whom?
3. A pod keeps restarting and `kubectl describe` shows exit code 137. What does that number tell you?
4. Why does `./deployctl "checkout api"` give `os.Args` a length of 2 and not 3?
5. Your script runs `export AWS_REGION=eu-west-1`. After it finishes, `echo $AWS_REGION` in your shell prints nothing. Why?
6. Why does a Go service that never calls `signal.Notify` lose in-flight requests during a rolling update?
7. Your tool updates `/etc/app/app.conf` with `os.WriteFile`. Why is writing a temp file in `/etc/app/` and renaming it over `app.conf` safer, and why not put the temp file in `/tmp`?

## Answers

1. The progress message was written to stdout, so it landed inside the file. Write messages to stderr (`fmt.Fprintln(os.Stderr, ...)`) and only the result to stdout.
2. Owner 6 = read and write; group 4 = read; others 0 = nothing. Handy for a log the ops group may read but nobody else may see.
3. 137 is 128 + 9: the process was killed by SIGKILL. In a pod that usually means it hit its memory limit and the OOM killer ended it (`kubectl describe` then shows `Reason: OOMKilled`), or it ignored SIGTERM until the grace period ran out.
4. The shell removes the quotes and passes `checkout api` as one argument. `os.Args[0]` is the program name, `os.Args[1]` is `checkout api`: two items.
5. The script ran as a child process with its own copy of the environment. Its change disappeared when it exited. A child can never change its parent's environment. (Running it with `source script.sh` executes it inside your own shell instead, so the change sticks.)
6. Kubernetes sends SIGTERM to stop the old pods. With no handler, Go's default is to exit immediately, so requests being served at that moment are cut off. Catching SIGTERM lets the service finish them first.
7. `os.WriteFile` empties the file and then writes it, so a reader or a crash in between sees a broken config. A rename is atomic: readers get the whole old file or the whole new one. Rename is only atomic within one filesystem, and `/tmp` is often a different one, so the temp file goes next to the target.

## Key terms

| Term | Meaning |
|---|---|
| kernel | the core of the OS that controls hardware and shares it out |
| user space | where normal programs run, without direct hardware access |
| system call | a request from a program to the kernel |
| process | a running program with its own memory, files and PID |
| PID / parent | a process's ID / the process that started it |
| thread | a line of execution inside a process, sharing its memory |
| absolute / relative path | a path from the root `/` / a path from the working directory |
| working directory | the folder relative paths are resolved from |
| permission bits | read, write, execute for owner, group and others |
| octal | base 8; how permissions like `0644` are written |
| umask | a per-process mask that removes permission bits from new files, e.g. `022` |
| atomic rename | replacing a file by renaming a finished temp file over it, so readers see all old or all new |
| file descriptor | the number a process uses for an open file or stream |
| stdin / stdout / stderr | fds 0, 1, 2: input, results, messages |
| pipe | the vertical bar between commands: stdout of one into stdin of the next |
| environment variable | a `NAME=value` setting copied to each child process |
| exit code | the 0–255 number a process returns; 0 is success |
| signal | a numbered notice from the kernel, like SIGTERM |
| OOM killer | the kernel feature that kills processes when memory runs out |
| namespaces / cgroups | kernel features that limit what a container can see / how much it can use |

## Watch

- [Operating Systems: Crash Course Computer Science #18](https://www.youtube.com/watch?v=26QPDBe-NB8) · CrashCourse · 14 min: a quick, visual history of why kernels exist and how they share the CPU and memory between programs.
- [Unix Pipeline (Brian Kernighan) - Computerphile](https://www.youtube.com/watch?v=bKzonnwoR2I) · Computerphile · 5 min: one of Unix's creators explains pipes and small tools joined by stdin and stdout.
- [Terminating with Grace](https://www.youtube.com/watch?v=Z_l_kE1MDTc) · Google Cloud Tech · 5 min: how Kubernetes stops a pod with SIGTERM, a grace period and then SIGKILL, and what your program should do.
- [Containers From Scratch • Liz Rice • GOTO 2018](https://www.youtube.com/watch?v=8fi7uSYlOdc) · GOTO Conferences · 43 min: Liz Rice builds a container in Go live, using processes, syscalls and the filesystem from this guide.

## Go deeper

- [Operating Systems: Three Easy Pieces](https://pages.cs.wisc.edu/~remzi/OSTEP/): a free, friendly textbook; the chapters on processes and the process API are a great next step.
- [The Linux Command Line](https://linuxcommand.org/tlcl.php): William Shotts' free book; the chapters on redirection, permissions and processes match this guide.
- [Standard streams](https://en.wikipedia.org/wiki/Standard_streams): the history of stdin, stdout and stderr, and why stderr exists at all.
- [signal(7)](https://man7.org/linux/man-pages/man7/signal.7.html): the Linux manual page listing every signal, its number and its default action.
- [Pod Lifecycle](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/): the official description of how Kubernetes terminates pods, grace period included.
