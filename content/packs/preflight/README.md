# Practice: preflight check CLI

Before a deploy, a CI job runs a quick *preflight check*: are the settings
sane, and is the host list right? You'll build `preflight`, one small tool
that uses everything from this module: subcommands, flags, streams, exit
codes, files, atomic writes and config precedence.

**New here:** `-` as a file name meaning stdin. The rest is practice.

## Learn

This is a practice exercise, so it only recaps what you already know.

### The tools you'll use

| Tool | Where you learned it | Used here |
|---|---|---|
| `Run(args, stdin, stdout, stderr, …) int` | *Standard streams: stdin, stdout, stderr and exit codes* | the tests call it with fakes |
| a command table + generated help | *Subcommands: a command table and generated help* | `Run` picks `check`, `init` or `version` |
| a `FlagSet` per command | *Flags: parsing options with a FlagSet* | `-config`, `-force`, `-env`, … |
| `os.Stat`, `os.ReadFile`, `os.Open`, `bufio.Scanner` | *Files: read, scan and write* | reading the config and the host list |
| temp file + rename | *Atomic writes: temp file and rename* | `init` writes the config safely |
| defaults < file < env < flags | *Config precedence: defaults, file, env and flags* | `Load` and `fs.Visit` |
| grep-style exit codes | *Standard streams: grep-style exit codes* | `check` returns 0, 1 or 2 |

As in *Config precedence: defaults, file, env and flags*, the environment
comes in as a `lookup` function: real code passes `os.LookupEnv`, the tests
pass a fake backed by a map.

### How preflight fits together

```text
preflight check -env prod
        │
        ▼
Run ── commands table ──► runCheck
                            │ 1. flags (FlagSet)
                            │ 2. Load: Defaults() → preflight.json → PREFLIGHT_* env
                            │ 3. typed flags win (fs.Visit)
                            │ 4. Validate
                            │ 5. CheckHosts(host list or stdin)
                            ▼
              stdout: one line per problem
              stderr: summary or error      exit 0 / 1 / 2
```

Everything lives in `exercise.go`, in task order:

| Piece | What it does | Status |
|---|---|---|
| `commands`, `writeUsage`, `runVersion` | the command table, help text, a finished example command | done |
| `Run` | picks the command | task 1 |
| `Settings`, `Defaults` | the three settings and their built-in values | done |
| `Validate` | rejects bad settings | task 2 |
| `runInit` | `preflight init` | task 3 |
| `WriteFileAtomic` | your atomic write from *Atomic writes: safely change one setting* | done |
| `CheckHosts` | finds problems in a host list | task 4 |
| `fileSettings`, `Load` | defaults < file < env; you declare `fileSettings` | task 5 |
| `checkFlags`, `runCheck` | `preflight check`; its flags are already defined | task 6 |

### Exit codes

| Code | `check` | `init` |
|---|---|---|
| `0` | no problems | file written |
| `1` | problems found | file exists, or the write failed |
| `2` | usage or config error | usage error |

As with grep, `1` means "the answer is no", and `2` means "I couldn't even
check". A broken config must never look like a bad host list.

### New: `-` means stdin

Many tools accept `-` where a file name goes, and read stdin instead:
`cat hosts.txt | preflight check -hosts -`. Both `*os.File` and `stdin` are
`io.Reader`s, so one variable can hold either:

```go
var src io.Reader = stdin
if name != "-" {
	f, err := os.Open(name)
	if err != nil {
		return err
	}
	defer f.Close()
	src = f
}
lines, err := countLines(src) // works on both
```

> **Why is it written like this?** `var src io.Reader = stdin` spells out the
> interface type for clarity; `src := stdin` works too, because `stdin` is
> already an `io.Reader`. The `-` convention comes from POSIX, not from Go.
> Never close stdin: only close what you opened.

## Your task

Your deploy pipeline will run `preflight check` before every rollout. The
tool's frame is in `exercise.go`. The tests call `Run`, or a command's
`run…` function directly with the words after the command name.

### 1. Dispatch the commands

**Goal:** finish `Run` so it runs the command named by `args[0]` from the
`commands` table.

No arguments and `help`, `-h` or `--help` already print the usage. Add the
rest:

- An unknown name: `preflight: unknown command "NAME"` on stderr, then the
  usage on stderr, return `2`.
- Otherwise call the command's `run` with the arguments **after** the name,
  plus the same `stdin`, `stdout`, `stderr` and `lookup`, and return its exit
  code.

`writeUsage(w)` writes the help text.

```text
$ preflight version
preflight 1.4.0
$ preflight deploy -env prod
preflight: unknown command "deploy"
usage: preflight <command> [flags]

commands:
  check    check the settings and the host list before a deploy
  init     write a config file with the default settings
  version  print the version
```

`preflight version now` must fail with `preflight: version takes no
arguments`: `runVersion` must not see the word `version`.

**Think it through:**
- Which two values does looking up a map key give you?
- Which slice expression drops the first element of `args`?

### 2. Validate the settings

**Goal:** `Validate` returns an error for the first bad setting, checking
`Env`, then `HostsFile`, then `MaxHosts`; `nil` when all are good.

| Setting | Bad when | Error message |
|---|---|---|
| `Env` | not `dev`, `staging` or `prod` | `env must be dev, staging or prod, got "production"` |
| `HostsFile` | empty `""` | `hosts_file must not be empty` |
| `MaxHosts` | below 1 | `max_hosts must be at least 1, got 0` |

`"production"` and `0` stand for the actual value; the env value is in
double quotes.

```go
Defaults().Validate()
// => nil

Settings{Env: "qa", HostsFile: "", MaxHosts: 0}.Validate()
// => error: env must be dev, staging or prod, got "qa"
```

**Think it through:**
- Three allowed words: a `switch` with one `case` can list them all.
- Which messages need `fmt.Errorf`, and which never change?

### 3. Write the init command

**Goal:** finish `runInit` so `preflight init` writes the default settings
to a JSON file, atomically.

- Flags: `-config PATH` (default `preflight.json`) and `-force` (a `bool`,
  default `false`). Handle a parse error like `runVersion` does.
- Any word left after the flags: `preflight: init takes no arguments` on
  stderr, return `2`.
- If the file exists and `-force` was not given: `preflight: PATH already
  exists; use -force to replace it` on stderr, return `1`, and leave the file
  alone.
- The contents are `Defaults()` from `json.MarshalIndent` with a two-space
  indent, plus a final newline. Save them with `WriteFileAtomic`. If either
  fails, print `preflight: ` and the error to stderr and return `1`.
- On success print `wrote PATH` to stdout and return `0`.

PATH is the `-config` value exactly as typed.

```text
$ preflight init
wrote preflight.json
$ cat preflight.json
{
  "env": "staging",
  "hosts_file": "hosts.txt",
  "max_hosts": 10
}
$ preflight init
preflight: preflight.json already exists; use -force to replace it
```

**Think it through:**
- `WriteFileAtomic` already asks "does this file exist?". Which call does it
  use, and which result means yes?
- Where does the `"env"` key name come from, so you don't type it?
- `MarshalIndent` returns bytes. How do you add one more byte, `'\n'`?

### 4. Check the host list

**Goal:** `CheckHosts(r, env, maxHosts)` reads one host per line from `r`
and returns a message for each problem, in order.

Trim spaces around each line and skip blank lines and lines starting with
`#`, as in *Files: read, scan and write*. `N` is the line number, counting
**every** line from 1, blank and comment lines included. Each host line gets
at most one message, checked in this order:

| Problem | Message |
|---|---|
| a space or tab inside the host | `line N: invalid host "web 2.prod"` |
| doesn't end in `.` + env | `line N: host "db-1.staging" is not in prod` |
| same host as an earlier good line | `line N: duplicate host "web-1.prod"` |

After the last line, add one more message if needed:

- no host lines at all: `no hosts`
- more host lines than `maxHosts` (every host line counts, bad ones too):
  `12 hosts, more than max_hosts 10`

Return `nil` problems when everything is fine. If reading fails, return
`nil` and the scanner's error.

```go
hosts := "# web tier\nweb-1.prod\n\nweb-1.prod\napi 1.prod\ncache-1.dev\n"
CheckHosts(strings.NewReader(hosts), "prod", 10)
// => []string{
//      `line 4: duplicate host "web-1.prod"`,
//      `line 5: invalid host "api 1.prod"`,
//      `line 6: host "cache-1.dev" is not in prod`,
//    }, nil
```

`web-1.preprod` is not in `prod`: the dot matters.

**Think it through:**
- Which data structure remembers "seen this host already"?
- You need two counters. Which counts lines, and which counts hosts?
- `fmt.Sprintf` builds a message without printing it.

### 5. Load the config layers

**Goal:** `Load(path, lookup)` starts from `Defaults()`, applies the JSON file
at `path`, then the environment, and returns the result.

| Setting | JSON key | Environment variable |
|---|---|---|
| `Env` | `env` | `PREFLIGHT_ENV` |
| `HostsFile` | `hosts_file` | `PREFLIGHT_HOSTS_FILE` |
| `MaxHosts` | `max_hosts` | `PREFLIGHT_MAX_HOSTS` (convert with `strconv.Atoi`) |

- A layer only overrides what it mentions. At the `TODO(5)` above `Load`,
  declare `fileSettings` with one pointer field per setting, as in
  *Config precedence: defaults, file, env and flags*: `{"max_hosts": 0}`
  sets `MaxHosts` to `0`, a missing key changes nothing. A variable set to
  `""` counts as set.
- A file that doesn't exist is skipped. Any other read error (such as `path`
  being a directory) is returned.
- Invalid JSON: return an error that starts with the path, a colon and a
  space: `preflight.json: unexpected end of JSON input`.
- A `PREFLIGHT_MAX_HOSTS` that isn't a whole number: return an error that
  mentions `PREFLIGHT_MAX_HOSTS`, such as `invalid PREFLIGHT_MAX_HOSTS "ten":
  …`.
- Don't validate here; `runCheck` does that after the flags.

```go
// preflight.json: {"env": "prod", "hosts_file": "/srv/hosts.txt", "max_hosts": 3}
// environment:    PREFLIGHT_MAX_HOSTS=25
Load("preflight.json", lookup)
// => Settings{Env: "prod", HostsFile: "/srv/hosts.txt", MaxHosts: 25}, nil
```

**Think it through:**
- Which error tells "no such file" apart from the rest? The banner in
  *Files: read, scan and write* used it.
- Which of `lookup`'s two results means "set"?

### 6. Write the check command

**Goal:** finish `runCheck`: build the settings from all four layers, check
the host list and report with grep-style exit codes.

The flags are parsed already: their values are in `cli`. After that, in
order:

1. `Load(cli.config, lookup)`. On error: `preflight: ` + the error on stderr,
   return `2`.
2. Every flag that was **typed** (`-env`, `-hosts`, `-max-hosts`) overrides
   its setting, even when its value is `""` or `0`.
3. `Validate`. On error: `preflight: ` + the error on stderr, return `2`.
4. Open `HostsFile`, or use `stdin` when it is `-`. On error: `preflight: `
   + the error, return `2`.
5. Run `CheckHosts`. On a read error: `preflight: reading hosts: ` + the
   error, return `2`.
6. Print each problem on its own line to **stdout**. If there were any,
   print `preflight: N problem(s) found` to stderr and return `1`. If not,
   print `preflight: ok for ENV` to stderr and return `0`.

A relative `hosts_file` is relative to the folder you run `preflight` in.
Below, `prod.txt` holds `web-1.prod`, `web-1.prod` and `db-1.staging`, one
per line.

```text
$ preflight check -env prod -hosts prod.txt
line 2: duplicate host "web-1.prod"
line 3: host "db-1.staging" is not in prod
preflight: 2 problem(s) found
$ echo $?
1
$ printf 'web-1.dev\n' | PREFLIGHT_ENV=dev preflight check -hosts -
preflight: ok for dev
$ preflight check -max-hosts 0
preflight: max_hosts must be at least 1, got 0
```

**Think it through:**
- Which `FlagSet` method visits only the flags that were typed?
- `-hosts -` and `PREFLIGHT_HOSTS_FILE=-` both mean stdin. After step 2,
  which one value do you check?
- Which stream gets the problems, and which gets the summary?

## Learn more

- [Command Line Interface Guidelines](https://clig.dev/)
- [POSIX Utility Conventions](https://pubs.opengroup.org/onlinepubs/9799919799/basedefs/V1_chap12.html) (guideline 13: `-` means stdin)
- [Package flag](https://pkg.go.dev/flag)
- [The Twelve-Factor App: Config](https://12factor.net/config)
