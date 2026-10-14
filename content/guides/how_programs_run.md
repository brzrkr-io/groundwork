# How a program runs

**In one sentence:** a program starts as text you write, a compiler turns it
into a file of machine instructions, and the operating system loads that file
into memory so the CPU can carry out the instructions one by one.

**Why it matters:** every deploy you'll ever do is this journey. A CI pipeline
compiles your code, packs the result into a container image, and Kubernetes
asks an operating system to run it. When something breaks, knowing which
stage you're in tells you where to look.

**You'll learn:**

- What the CPU, memory and the operating system each do when a program runs.
- How Go source code becomes a binary, and what a process is.
- Why the compiler rejects code before it runs (compile time vs run time).
- How Go differs from Python and bash (compiled vs interpreted).
- What `go build`, `go run` and `go test` do, and what `package main` is for.
- How programs talk to the outside world: stdin, stdout, stderr and exit codes.

## The big picture

Here is the whole journey of a small deploy tool, from your editor to a
running program:

```text
 you write        the compiler       a file on disk     the OS loads it
┌──────────┐     ┌────────────┐     ┌────────────┐     ┌────────────┐
│ main.go  │ ──▶ │ go build   │ ──▶ │ deploybot  │ ──▶ │ a process  │
│ (text)   │     │ checks and │     │ (machine   │     │ running in │
│          │     │ translates │     │ code)      │     │ memory     │
└──────────┘     └────────────┘     └────────────┘     └────────────┘
 source code      compile time       the binary         run time
```

The rest of this guide walks through each box. Keep this picture in mind:
most confusion about "why didn't my code work" comes from mixing up two of
these stages.

## The machine: CPU and memory

### The CPU follows instructions

The *CPU* (central processing unit) is the chip that does the work. It can
only do very small things: add two numbers, copy a number from one place to
another, compare two numbers, jump to a different instruction.

Each of those small things is an *instruction*. A program, as the CPU sees
it, is a long list of them. The CPU fetches one, does it, and moves to the
next, billions of times a second.

### Memory holds what's running, disk keeps it

| | Memory (RAM) | Disk |
|---|---|---|
| Holds | Running programs and the data they're using | Files: binaries, logs, configs |
| Speed | Very fast | Much slower |
| When the power goes off | Everything is lost | Files stay |

A binary sits on disk doing nothing. To run it, the operating system copies
it into memory, because the CPU can only work on what's in memory.

This is why a pod has a *memory limit* in Kubernetes: it caps how much RAM
the running program may use. Its image size is about disk, which is a
different thing.

## Source code is for people, machine code is for the CPU

You write *source code*: text a person can read.

```go
total := used + free
```

The CPU can't read that. It needs *machine code*: instructions encoded as
numbers. Written out by hand, the line above becomes something like this
(simplified, for illustration):

```text
load   the value of used  into register A
load   the value of free  into register B
add    register B to register A
store  register A into the place for total
```

A *register* is a tiny storage slot inside the CPU itself. Real machine code
is just bytes, not words. Nobody wants to write that, which is why compilers
exist.

Machine code is specific to a kind of CPU. An Intel or AMD laptop uses one
instruction set (*amd64*), while an Apple Silicon Mac or an AWS Graviton
server uses another (*arm64*). A binary built for one won't run on the other.

## The compiler checks, then translates

A *compiler* is a program that reads source code and writes machine code.
Go's compiler runs when you type `go build`. It does two jobs:

1. **Checks** the whole program: every name exists, every value has the
   right type, every brace is closed.
2. **Translates** it into machine code and writes the binary.

If the check fails, there is no binary. Nothing runs at all.

### Compile time vs run time

*Compile time* is when the compiler reads your code. *Run time* is later,
when the binary is actually executing. Some mistakes are caught at compile
time, others only show up at run time.

```go
package main

import "fmt"

func main() {
	var port int
	port = "8080"
	fmt.Println("port: " + port)
}
```

This program never runs. `go build` stops with:

```text
./main.go:7:9: cannot use "8080" (untyped string constant) as int value in assignment
./main.go:8:14: invalid operation: "port: " + port (mismatched types untyped string and int)
```

Step by step: the compiler sees `port` declared as an `int` (a whole
number). On line 7 it's given text in quotes, which is a string, not a
number. On line 8 text is joined to a number with `+`, which Go doesn't
allow. Both problems are reported before a single instruction runs.

| Caught at compile time | Only caught at run time |
|---|---|
| A misspelled name (`undefined: hostname`) | The config file doesn't exist |
| Text where a number belongs | The database is down |
| A missing `}` | Dividing by a number that turns out to be 0 |
| A variable you never use | A wrong answer from correct-looking code |

**Why this is good news:** a compile error is the cheapest bug you'll ever
fix. It's found on your laptop in a second, not at 3 a.m. in production. The
*Reading compiler errors* exercise teaches you to read these messages.

## Compiled vs interpreted: Go, Python and bash

Not every language works like Go.

- **Compiled** (Go, Rust, C): the compiler turns everything into machine code
  ahead of time. You ship the binary. The machine that runs it doesn't need
  Go installed.
- **Interpreted** (bash, and Python in practice): you ship the source code.
  Another program, the *interpreter*, reads it and carries it out while it
  runs.

| | Go | Python | bash |
|---|---|---|---|
| What you ship | One binary | `.py` files | `.sh` files |
| Needs installed to run | Nothing extra | A Python interpreter and libraries | bash and every command it calls |
| A misspelled name is found | At compile time | When that line runs | When that line runs, or never (see below) |

That last row matters in ops. A bash deploy script with a typo on line 80
runs lines 1 to 79 first, then hits the typo halfway through a deploy. A Go
tool with the same typo never builds.

bash is even sneakier with variables. A misspelled variable like `$HOSNAME`
isn't an error at all: bash quietly swaps in empty text and carries on.
(The *Shell variables* exercise shows the `set -u` line that turns this
into an error.)

(Python does check its *syntax* when it loads a file, and turns your code
into an in-between form called bytecode first. But a misspelled variable
name still only fails when that line runs.)

## The operating system runs your program

The *operating system* (OS: Linux, macOS, Windows) is the program that
manages the machine for every other program. When you run a binary, the OS:

1. Finds the file on disk and copies its instructions into memory.
2. Gives it its own private area of memory to work in.
3. Connects its input and output (more on that below).
4. Tells the CPU to start at the program's first instruction.
5. Shares the CPU between it and every other running program.

A running program is called a *process*. The binary is the recipe; the
process is the cooking. You can start the same binary three times and get
three separate processes, each with its own memory and its own *process ID*
(PID), a number the OS uses to tell them apart.

```bash
ps aux    # list every process running on this machine
```

A container, like a Kubernetes pod's container, is still just a process. The
OS fences it off so it sees its own files and network. You'll go deeper in
*The operating system for programmers* guide.

## A package vs a program

Go code is organised in *packages*: folders of `.go` files that start with
the same `package` line. There are two kinds.

| | Library package | Program |
|---|---|---|
| First line | `package hello` (any name) | `package main` (exactly) |
| Has `func main()` | No | Yes, exactly one |
| `go build` makes | No binary, just checks it | A binary you can run |

Every exercise in *Basics* is a library package. Its functions are run by
the tests, not by you. A program needs `package main` and a function called
`main`, which is where the process starts:

```go
package main

import "fmt"

func main() {
	fmt.Println("deploying api to production")
}
```

Save this as `main.go` in a folder with a `go.mod` file (make one with
`go mod init hello`), then run `go run .`. It prints:

```text
deploying api to production
```

Step by step: the OS starts the process, Go sets itself up, and then calls
`main`. `main` calls `fmt.Println`, which writes the text plus a newline to
the screen. When `main` reaches its closing `}`, the program ends.

## go run, go build and go test

The `go` command is your toolbox. These three do most of the work:

| Command | What it does | What's left afterwards |
|---|---|---|
| `go build` | Checks and compiles the program into a binary | A binary named after the module (`hello`) |
| `go run .` | Compiles the program out of sight (in Go's cache), then runs it | Only the output; no binary in your folder |
| `go test` | Compiles the package plus its `_test.go` files into a test binary and runs it | A pass/fail report |

```bash
go build      # makes ./hello
./hello       # runs it: deploying api to production
```

`go run` is for trying things out. `go build` is what a pipeline does,
because the binary is the thing you ship. `go test` is what Groundwork runs
every time you save an exercise. It also runs a few of `go vet`'s checks,
which look for code that compiles but is almost certainly wrong.

Even this tiny program's binary is about 2 MB. Every Go binary carries the
Go *runtime*: code that cleans up memory you've stopped using and juggles
the program's tasks for you. That's also
why it runs on a machine with no Go installed.

## Talking to the world: stdin, stdout and stderr

Every process starts with three connections already open, called the
*standard streams*:

```text
                  ┌─────────────┐ ──▶ stdout (1): the results
 stdin (0) ─────▶ │  a process  │
  the input       └─────────────┘ ──▶ stderr (2): errors and progress
```

- **stdin** (standard input): where input comes from, usually the keyboard
  or another program.
- **stdout** (standard output): where results go. `fmt.Println` writes here.
- **stderr** (standard error): where error messages go.

By default both outputs go to your terminal. The shell lets you send them
elsewhere. `>` sends stdout to a file:

```bash
./hello > out.txt   # the screen shows nothing; out.txt has the line
go build > out.txt  # compile errors still appear: they go to stderr
```

Why two outputs? So errors don't get mixed into results. When a script runs
`kubectl get pods -o name > pods.txt`, you want only pod names in the file,
while a warning still reaches your screen. The pipe `|` joins one program's
stdout to the next one's stdin:

```bash
kubectl get pods | grep CrashLoopBackOff
```

## Exit codes: did it work?

When a process ends, it hands the OS one small number, its *exit code*.
**0 means success. Anything else means failure.** The meaning of each
non-zero number is up to the program.

```bash
ls /nope
echo $?    # $? holds the last command's exit code
```

```text
ls: cannot access '/nope': No such file or directory
2
```

This tiny number runs the whole automation world:

- A CI pipeline step fails when its command exits non-zero.
- `go build` exits 1 when there are compile errors, so a broken commit never
  becomes an image.
- Kubernetes records a container's exit code; a container that keeps
  exiting and being restarted ends up in `CrashLoopBackOff`.
- In bash, `cmd1 && cmd2` runs `cmd2` only if `cmd1` exited 0.

A Go program whose `main` returns normally exits 0. Choosing a different
code, and writing to stderr on purpose, comes in *Standard streams: stdin,
stdout, stderr and exit codes*, in the *CLI tools and the OS* module.

## Putting it together: from git push to a running pod

Here's the journey from the top of this guide, as a real pipeline runs it:

```text
git push
   │
   ▼
CI runner (a fresh Linux machine)
   ├─ go test ./...         exit 0?  no ──▶ pipeline stops, you're told
   ├─ CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o deploybot
   │                        one binary, built for the servers' CPU
   ├─ docker build          copy the binary into a small image
   └─ docker push           store the image in a registry
   │
   ▼
Kubernetes pulls the image and starts a container:
the OS loads deploybot into memory, and the CPU runs main()
```

The image can be tiny, because a Go binary needs no interpreter:

```text
FROM gcr.io/distroless/static
COPY deploybot /deploybot
ENTRYPOINT ["/deploybot"]
```

`GOOS` and `GOARCH` tell the compiler which OS and CPU to build for. That's
how you build for arm64 Graviton nodes from an amd64 laptop.
`CGO_ENABLED=0` makes sure the binary needs nothing else from the image.

## Common mistakes

- **Thinking code "runs" while the compiler complains.** If there's a
  compile error, nothing ran. Fix the first error, save, and read the list
  again.
- **Confusing the binary with the process.** Replacing the binary on disk
  doesn't change a process that's already running. Only a restart loads the
  new one.
- **Writing errors to stdout.** Scripts that read your output get garbage
  mixed in. Errors belong on stderr.
- **Ignoring exit codes.** A script that runs `go test` but doesn't stop on a
  non-zero exit will happily deploy broken code.
- **Building for the wrong CPU.** An amd64 binary on an arm64 node fails with
  `exec format error`. Set `GOARCH` to match the nodes.
- **Expecting `package hello` to run.** Only `package main` with a
  `func main()` makes a runnable program.

## Check yourself

1. You run `go build` and get `undefined: hostname`. Did any of your code
   run? Why or why not?
2. What's the difference between the file `deploybot` on disk and a
   `deploybot` process?
3. A bash script and a Go program both contain a misspelled variable name
   near the end. When does each one find out?
4. A command prints results and also an error. You run it with
   `> out.txt`. Which text ends up in the file, and which on your screen?
5. A CI step runs a command that exits with code 3. What does the pipeline
   do, and why?

## Answers

1. No. `undefined: hostname` is a compile error, so the compiler stopped
   before writing a binary. With no binary, there's nothing for the OS to
   load or the CPU to run.
2. The file is the binary: machine code stored on disk, doing nothing. A
   process is that binary loaded into memory and running, with its own
   memory and PID. One file can be running as several processes at once.
3. The Go program finds out at compile time, before anything runs, because
   the compiler checks every name. The bash script doesn't find out at all
   by default: bash treats the unknown variable as empty text and keeps
   going. Only with `set -u` does it fail, and then only at run time, when
   it reaches that line after running everything before it.
4. The results go to stdout, so they land in `out.txt`. The error goes to
   stderr, which `>` doesn't redirect, so it still appears on your screen.
5. It marks the step as failed and, usually, stops the pipeline. Any exit
   code other than 0 means failure, whatever the number.

## Key terms

| Term | Meaning |
|---|---|
| CPU | The chip that carries out instructions, one tiny step at a time. |
| Memory (RAM) | Fast storage for running programs and their data; wiped at power-off. |
| Instruction | One tiny CPU operation, such as "add these two numbers". |
| Source code | The text of a program, written for people to read. |
| Machine code | Instructions encoded as numbers for one kind of CPU. |
| Compiler | A program that checks source code and translates it into machine code. |
| Binary | The file of machine code the compiler produces. |
| Compile time | When the compiler checks and translates your code. |
| Run time | When the program is actually running. |
| Interpreter | A program that reads source code and carries it out as it runs. |
| Operating system (OS) | The program that manages the machine and runs other programs. |
| Process | A running instance of a program, with its own memory and PID. |
| Package | A folder of Go files sharing one `package` line. |
| Runtime | Code Go puts in every binary to manage memory and tasks. |
| Standard streams | stdin, stdout and stderr: a process's input, output and error output. |
| Exit code | The number a process returns when it ends: 0 for success. |

## Watch

- [Go in 100 Seconds](https://www.youtube.com/watch?v=446E-r0rXHI) · Fireship · 2 min: a fast, fun overview of what Go is and why it compiles to one fast program file.
- [The Central Processing Unit (CPU): Crash Course Computer Science #7](https://www.youtube.com/watch?v=FZGugFqdr60) · CrashCourse · 12 min: shows, step by step, how the CPU fetches and runs one tiny instruction at a time from memory, the picture behind "the CPU follows instructions".
- [How do computers read code?](https://www.youtube.com/watch?v=QXjU9qTsYCc) · Frame of Essence · 12 min: an animated walk from the code a person types down to the instructions the CPU actually runs.
- [AT&T Archives: The UNIX Operating System](https://www.youtube.com/watch?v=tc4ROCJYbm0) · AT&T Tech Channel · 27 min: a classic 1982 film from Bell Labs; watch Brian Kernighan chain small programs into a pipeline, the idea behind stdin and stdout.

## Go deeper

- [Tutorial: Get started with Go](https://go.dev/doc/tutorial/getting-started)
  — write, run and build your first `package main` program yourself.
- [How to Write Go Code](https://go.dev/doc/code) — how modules, packages,
  `go build` and `go install` fit together.
- [Operating Systems: Three Easy Pieces, chapter 2: Introduction](https://pages.cs.wisc.edu/~remzi/OSTEP/intro.pdf)
  — a free, friendly book; this chapter shows what the OS does when a program
  runs.
- [Wikipedia: Standard streams](https://en.wikipedia.org/wiki/Standard_streams)
  — where stdin, stdout and stderr came from, including why stderr was
  invented.
