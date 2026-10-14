# The shell and the command line

**In one sentence:** the *shell* is a program that reads a line you type, rewrites it (variables, quotes, wildcards) into a list of words, and starts the program those words name; bash is the shell you'll meet on almost every server and in CI.

**Why it matters:** at 3 a.m. during an incident there is no IDE, just a prompt on a server. Deploy scripts, CI jobs, Dockerfiles and Kubernetes probes are full of shell commands. Knowing how the shell rewrites a line before running it is the difference between `rm "$build_dir"/*` and deleting the wrong thing.

**You'll learn:**
- what the shell does to a line before it runs anything, and how quotes change that
- variables, arguments and `$(command)`, and why you write `"$var"` with double quotes
- exit status, `&&`, `||`, `if`, loops and functions in bash
- how to write a script that stops on the first error (`set -euo pipefail`)
- the text toolbox (grep, sort, uniq, cut, awk, sed, find, xargs) and how to chain it into a pipeline
- when a script has outgrown bash and should become a Go program

## Terminal, shell, bash

Open a terminal and type `ls`. Three different programs are involved.

| Piece | What it is | Examples |
|---|---|---|
| *terminal* | the window: draws text, sends your keystrokes | GNOME Terminal, iTerm2, a VS Code panel |
| *shell* | reads your line, works out the words, starts programs | bash, zsh, sh |
| command | the program that does the work | `ls`, `grep`, `kubectl`, your Go binary |

The shell is an ordinary program whose job is starting other programs. *bash* (the Bourne Again SHell) is the default on most Linux servers and CI runners, so this guide uses it. zsh, the macOS default, differs in places (it doesn't split unquoted variables), so run scripts with bash.

Processes, file descriptors, redirection (`>`, `2>`), pipes (`|`) and exit codes belong to the kernel; the shell only sets them up. *The operating system for programmers*, the guide before this one, explains them; this guide shows how you type them.

## Commands, arguments and options

```bash
kubectl get pods -n payments --watch
```

The first word is the *command*, the program to run. Every word after it is an *argument*, passed to the program as a list of strings (in Go, `os.Args`).

By convention, arguments that start with a dash are *options* (also called flags):

| Form | Example | Meaning |
|---|---|---|
| short option | `ls -l` | one letter, one dash |
| short options combined | `ls -la` | same as `ls -l -a` |
| long option | `ls --all` | a whole word, two dashes |
| option with a value | `-n payments`, `--namespace=payments` | the next word (or after `=`) is the value |
| `--` | `rm -- -oops.log` | "no more options": everything after is a plain argument |

**The program decides what its options mean, not the shell.** `-n` is "namespace" to kubectl and "line numbers" to grep. Check with `man grep` or `grep --help`. You'll practise this in *Shell commands and quoting*.

## Files and folders

Most commands work on files, and every file has a path. A path that starts with `/` is *absolute* and means the same thing from anywhere. Any other path is *relative* to the folder you're in (`pwd` prints it): `logs/app.log`, `./deploy.sh` (`.` is this folder), `../shared/app.yaml` (`..` is the folder above).

| Command | Does | Watch out |
|---|---|---|
| `mkdir -p a/b/c` | make a folder, and any missing parents | without `-p`, a missing parent is an error |
| `cp src dest` | copy a file | `cp -r` for a folder |
| `mv src dest` | move or rename | silently replaces an existing `dest` |
| `rm file` | delete a file | no undo, no trash can; `rm -r` deletes a whole folder |
| `ls -a` | list, including hidden files | |

A name that starts with a dot (`.env`, `.gitignore`, `.git/`) is *hidden*: `ls` and the wildcard `*` skip it, `ls -a` shows it. Config files hide this way so they don't clutter listings, which is also how secrets in `.env` get forgotten. You'll practise all of this in *Files and folders: paths, mkdir, cp, mv and rm*.

## What the shell does to a line

Quotes only make sense once you see the steps. Before running anything, bash rewrites your line in this order:

```text
  rm $file *.tmp
       │
  1. split the line into words at spaces   →  rm | $file | *.tmp
     (quoted text stays one word)
  2. expand $variables and $(commands)     →  rm | release notes.txt | *.tmp
  3. split unquoted results at spaces      →  rm | release | notes.txt | *.tmp
  4. expand unquoted globs (* ? [ ])       →  rm | release | notes.txt | a.tmp | b.tmp
  5. remove the quote characters
  6. run the first word with the rest as arguments
```

Steps 3 and 4 are where bugs come from: a variable's value gets split again and wildcard-expanded, unless it is inside double quotes. Splitting the result of an expansion is called *word splitting*; expanding `*` into file names is called *globbing*.

### Three kinds of quoting

| You write | `$var` expanded? | split at spaces? | `*` expanded? | Use it for |
|---|---|---|---|---|
| `$var` (no quotes) | yes | **yes** | **yes** | almost never |
| `"$var"` (double) | yes | no | no | every variable, every time |
| `'$var'` (single) | no, stays literally `$var` | no | no | fixed text, regexes, awk programs |

Quotes don't end up in the program's arguments. They are instructions to the shell, removed in step 5.

### A file name with a space

```bash
file="release notes.txt"
rm $file      # rm gets 2 arguments: "release" and "notes.txt"
rm "$file"    # rm gets 1 argument: "release notes.txt"
```

1. `file="release notes.txt"` stores one string with a space in it. The quotes keep the space inside the value.
2. `rm $file`: step 2 replaces `$file` with its value, step 3 splits it at the space. rm tries to delete two files that don't exist, or worse, two that do.
3. `rm "$file"`: the double quotes stop step 3. rm receives exactly one argument.

### An empty variable

```bash
env=""
[ $env = prod ]      # bash: [: =: unary operator expected
[ "$env" = prod ]    # false, as you meant
```

Unquoted, the empty value disappears completely, so `[` sees only `= prod` and fails. Quoted, it's an empty string and the comparison works.

Empty variables can be dangerous too. If `$build_dir` is empty, `rm -rf $build_dir/*` becomes `rm -rf /*`. Write `rm -rf "${build_dir:?}"/*`: `:?` stops the script with an error when the variable is empty.

**Rule:** put double quotes around every `$`: `"$var"`, `"$1"`, `"$@"`, `"$(cmd)"`. Leave them off only when you can say why. You'll fix exactly these bugs in *Shell variables: fix the quoting bugs*.

## Variables and `$`

```bash
service=api                  # no spaces around =
replicas=3
echo "deploying ${service}_v2 with $replicas replicas"
```

- **No spaces around `=`.** `service = api` runs a command called `service` with the arguments `=` and `api`.
- **`$name` reads a variable.** Write `${name}` when a letter, digit or `_` follows it: `"${service}_v2"` reads `service`, but `"$service_v2"` looks up a variable called `service_v2`.
- **Every value is a string.** `replicas=3` stores the text `3`.

### Special variables in scripts

| Variable | Meaning | In `./deploy.sh api prod` |
|---|---|---|
| `$0` | the script's name | `./deploy.sh` |
| `$1`, `$2`, ... | positional arguments | `api`, `prod` |
| `$#` | how many arguments | `2` |
| `"$@"` | all arguments, each kept as one word | `api` `prod` |
| `$?` | exit status of the last command | `0` if it worked |

A default for a missing argument: `env="${2:-staging}"` uses `$2` if it's set and not empty, otherwise `staging`.

### Command substitution

`$(command)` runs a command and puts its output into the line. This is *command substitution*.

```bash
today="$(date +%F)"            # date +%F prints e.g. 2026-10-13
echo "writing backup-${today}.tar.gz"
```

### export and the environment

A plain variable exists only inside this shell. The *environment* is the set of variables that every program the shell starts receives a copy of. `export` puts a variable there:

```bash
region=eu-west-1             # plain: only this shell sees it
export AWS_REGION=eu-west-1  # exported: terraform, aws and your Go program see it
LOG_LEVEL=debug ./deploy.sh  # set for this one command only
env | grep AWS_              # env prints the whole environment
```

A child gets a **copy**, so nothing it changes comes back to the shell that started it. `unset NAME` removes a variable. *Unset* and *empty* are different: `NAME=` sets it to the empty string. `${NAME:-default}` uses the default in both cases, `${NAME-default}` only when it's unset, and `set -u` stops the script only on an unset one. You'll practise all of this in *Shell variables*.

## Exit status, `&&` and `||`

Every command ends with an *exit status*: 0 means success, anything else (1–255) means failure. The shell keeps the last one in `$?`. The details are in *The operating system for programmers*.

```bash
grep -q ERROR app.log
echo $?      # 0 if ERROR was found, 1 if not, 2 if app.log is missing
```

`&&` and `||` chain commands by exit status:

| You write | Runs the second command when | Typical use |
|---|---|---|
| `a && b` | `a` succeeded | `go test ./... && go build` |
| `a \|\| b` | `a` failed | `cd /srv/app \|\| exit 1` |
| `a; b` | always | same as writing `a` and `b` on two lines |

```bash
go test ./... && docker build -t api . && docker push api
```

The image is pushed only if tests and the build both pass. Practise in *Exit codes, && and ||*.

## Redirection and here-docs

*The operating system for programmers* shows what redirection does to a process's streams. In bash you write it at the end of a command:

```bash
echo "replicas: 3" > app.yaml        # stdout into a file, replacing it
echo "port: 8080" >> app.yaml        # stdout added to the end
./deploy.sh 2> errors.txt            # stderr into a file
./deploy.sh > deploy.log 2>&1        # both into one file: > first, then 2>&1
grep -c ERROR < app.log              # stdin from a file
rm -f old.pid 2> /dev/null           # throw errors away
```

**`>` empties the file before the command even starts.** `sort app.log > app.log` leaves an empty file, because the shell truncated it before `sort` read a line. Write to a new file, then `mv` it over the old one.

A *here-doc* feeds several lines to a command's stdin, written right in the script. With `cat >`, it writes a whole file:

```bash
cat > config/db.yaml <<'EOF'
host: db-new.internal
port: 5432
EOF
```

Everything up to the line holding only `EOF` becomes the input. Quote the first `'EOF'` and bash leaves `$` alone, so the file gets exactly what you typed; unquoted, `$variables` are expanded first. Practise in *Redirection: >, >>, 2> and here-docs*.

## Conditions: `if` and test

`if` runs a command and checks its exit status. 0 means "true".

```bash
if grep -q "panic" app.log; then
  echo "found a panic, paging on-call"
fi
```

Most conditions use `[[ ... ]]`, bash's test syntax:

```bash
if [[ -f /etc/myapp/config.yaml ]]; then
  echo "config found"
elif [[ -z "$CONFIG_URL" ]]; then
  echo "no config and no CONFIG_URL" >&2
  exit 1
fi
```

The spaces inside the brackets are required: `[[-f` is an error.

| Test | True when |
|---|---|
| `-f path` | path is a regular file |
| `-d path` | path is a directory |
| `-z "$s"` | the string is empty |
| `-n "$s"` | the string is not empty |
| `"$a" == "$b"` | the strings are equal |
| `"$a" != "$b"` | the strings differ |
| `$x -eq $y`, `-ne`, `-lt`, `-gt`, `-le`, `-ge` | numeric comparison |

**Strings vs numbers is a trap.** `==`, `<` and `>` compare text, letter by letter. `[[ 10 > 9 ]]` is false, because the character `1` sorts before `9`. Use `-gt` for numbers: `[[ "$errors" -gt 9 ]]`.

**`[ ]` vs `[[ ]]`.** `[` is the older command, also called `test`, and works in every POSIX shell. `[[` is bash-only but safer: it's part of bash's own syntax, so it doesn't word-split variables. In bash scripts, use `[[ ]]`, and still quote. Practise in *if and test*.

## Loops

**Over arguments or a list:**

```bash
for host in "$@"; do
  ssh "$host" uptime
done
```

`"$@"` gives each argument as one word, even if it contains spaces. Unquoted `$@` would split them.

**Over an array.** Bash arrays hold a list of strings:

```bash
services=(api worker "billing sync")
for s in "${services[@]}"; do
  systemctl restart "$s"
done
```

`"${services[@]}"` expands to exactly three words. Without the quotes, `billing sync` becomes two.

**Over the lines of a file:**

```bash
while IFS= read -r line; do
  echo "checking $line"
done < hosts.txt
```

`< hosts.txt` feeds the file to the loop. `read -r line` reads one line into `line`; `-r` keeps backslashes, and `IFS=` keeps leading and trailing spaces. Learn the line as one unit. Practise in *Shell loops*.

## Functions

```bash
log() {
  echo "$(date +%T) $*" >&2
}

current_version() {
  local file="$1"
  grep '^version:' "$file" | cut -d' ' -f2
}

version="$(current_version chart.yaml)"
log "deploying version $version"
```

- **Arguments** arrive as `$1`, `$2`, `"$@"`, just like a script's. `$*` (used in `log`) joins all of them into one string with spaces.
- **`local`** keeps a variable inside the function. Without it, every variable is global and can overwrite one the caller uses.
- **Return a result by printing it**, and capture it with `$(fn)`. `return` only sets the exit status (0–255): use it for success or failure, not for data.

Practise in *Shell functions*.

## Scripts

```bash
#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 1 ]]; then
  echo "usage: $0 <service>" >&2
  exit 2
fi
service="$1"

echo "restarting $service" >&2
systemctl restart "$service"
systemctl is-active --quiet "$service"
echo "ok"
```

The first line is the *shebang*. When you run `./restart.sh`, the kernel reads it and runs the file with the program it names; `/usr/bin/env bash` means "find bash on the PATH". `chmod +x restart.sh` makes the file executable. Or skip both and run `bash restart.sh`, as the exercises do.

A script that works by hand can still fail under cron or CI, which start it with their own `PATH` and working directory: `command not found`, or a relative path that points nowhere. `echo "$PATH"`, `command -v tool` and `pwd` tell you which; you'll diagnose and fix one in *Processes and permissions: an incident on the box*.

### `set -euo pipefail`

By default bash keeps going after a failed command. A deploy script that fails at "build" then carries on to "push" and "restart". These options make the script fail fast, like returning an error in Go:

| Option | Effect | Without it |
|---|---|---|
| `-e` | exit when a command fails | the script runs on after errors |
| `-u` | an unset variable is an error | `$sevice` (typo) quietly becomes empty |
| `-o pipefail` | a pipeline fails if any stage fails | only the last stage counts, so `grep ERROR missing.log \| wc -l` "succeeds" and prints `0` |

`-e` is not magic: it ignores failures inside an `if` condition or on the left of `&&` and `||`. Still check what matters.

It also stops on "failures" you may not expect. `grep` exits 1 when nothing matches, so on a clean log `count="$(grep -c ERROR app.log)"` ends the script, and so does `grep ERROR app.log | wc -l` under `pipefail`. When "no matches" is fine, say so: `count="$(grep -c ERROR app.log || true)"` still sets `count` to `0`.

### stdout vs stderr

Results go to stdout; messages and errors go to stderr with `>&2`. Then `./report.sh > report.csv` gets only the data, and the warnings still show on screen. (What `>&2` really does is in *The operating system for programmers*.)

## The text toolbox

Logs, config and command output are lines of text. Unix gives you small tools that each do one thing to lines:

| Tool | Does | Example |
|---|---|---|
| `grep` | keep lines matching a pattern | `grep -i 'timeout' app.log` |
| `sed` | edit lines (find and replace) | `sed 's/staging/prod/g' values.yaml` |
| `awk` | pick and compute with columns | `awk '$9 >= 500 {print $7}' access.log` |
| `cut` | cut out columns by delimiter | `cut -d, -f1,3 hosts.csv` |
| `sort` | sort lines (`-n` numbers, `-r` reverse) | `sort -rn counts.txt` |
| `uniq` | merge adjacent equal lines (`-c` counts) | `sort \| uniq -c` |
| `tr` | translate or delete characters | `tr 'a-z' 'A-Z'` |
| `wc` | count lines, words, bytes | `wc -l app.log` |
| `head` / `tail` | first / last lines (`tail -f` follows) | `tail -n 50 app.log` |
| `find` | list files by name, age, size | `find /var/log -name '*.gz'` |
| `xargs` | turn input lines into arguments | `... \| xargs rm` |

Single quotes around patterns are deliberate: `'$9 >= 500'` must reach awk untouched, not be expanded by the shell.

A *pipeline* (`a | b | c`) connects each command's stdout to the next one's stdin. You'll use these tools in *Grep and pipes*, *Sort and uniq: top-N reports*, *Columns with cut and awk* and *Editing text with sed*.

## Building a pipeline step by step

During an incident someone asks: "what are the top 3 errors in this log?" Here is `app.log`:

```text
2026-10-13 09:14:02 ERROR db-timeout order=1182
2026-10-13 09:14:03 INFO request-ok order=1183
2026-10-13 09:14:05 ERROR payment-declined order=1184
2026-10-13 09:14:09 ERROR db-timeout order=1185
2026-10-13 09:14:11 WARN slow-request order=1186
2026-10-13 09:14:12 ERROR disk-full order=1187
2026-10-13 09:14:15 ERROR db-timeout order=1188
2026-10-13 09:14:20 ERROR payment-declined order=1189
```

The answer:

```bash
grep ERROR app.log | cut -d' ' -f4 | sort | uniq -c | sort -rn | head -3
```

Build it one stage at a time, and look at the output after each one:

| Stage | Output |
|---|---|
| `grep ERROR app.log` | the 6 ERROR lines |
| `\| cut -d' ' -f4` | the 4th space-separated field: `db-timeout`, `payment-declined`, `db-timeout`, `disk-full`, `db-timeout`, `payment-declined` |
| `\| sort` | equal lines next to each other: `db-timeout` ×3, `disk-full`, `payment-declined` ×2 |
| `\| uniq -c` | `3 db-timeout`, `1 disk-full`, `2 payment-declined` |
| `\| sort -rn` | by count, biggest first: `3 db-timeout`, `2 payment-declined`, `1 disk-full` |
| `\| head -3` | the first 3 lines |

```text
      3 db-timeout
      2 payment-declined
      1 disk-full
```

**`uniq` needs `sort` first.** It only merges lines that are next to each other. **Build pipelines incrementally**: run the first stage, check, add the next. You'll write a full version in *Practice: log triage report*.

### find and xargs safely

Delete app logs older than 14 days:

```bash
find /var/log/myapp -name '*.log' -mtime +14 -print0 | xargs -0 rm --
```

1. `'*.log'` is quoted so the shell doesn't glob it; find needs the pattern itself.
2. `-mtime +14` keeps files last modified more than 14 days ago. find counts whole days and drops the fraction, so a file 14½ days old doesn't match yet: in practice `+14` means "15 days or older".
3. `-print0` ends each name with a zero byte instead of a newline, and `xargs -0` splits on that byte. File names can contain spaces and even newlines, but never a zero byte.
4. `xargs` runs `rm --` with the names as arguments. Run it with `echo` instead of `rm` first. (GNU find can also delete on its own with `-delete`.)

Practise in *Find and xargs: clean up old logs*.

## When to stop using bash

Bash is great at gluing programs together. It is poor at everything else. Rewrite in Go when you see these signs:

| Sign | Why Go is better |
|---|---|
| over ~100 lines | bash has no types and few ways to structure code |
| you need maps, structs or nested lists | bash arrays hold only strings, one level deep |
| you parse JSON or YAML | Go has `encoding/json`; in bash you'd need `jq` installed |
| careful error handling matters | Go errors are values you must check (see *Errors*) |
| it needs tests | Go has `go test`; bash testing is awkward |
| it must run on many machines | one static binary; no "which bash version is this?" |

A good rule: a short script that runs commands in order stays bash. A script that makes decisions about data becomes Go. You'll build those tools in *CLI tools and the OS* and *Data formats and parsing*.

## Common mistakes

- **Unquoted variables.** `rm $file` splits names with spaces and globs `*`. Write `"$file"`.
- **Spaces around `=`.** `name = api` runs a command called `name`. Write `name=api`.
- **Comparing numbers with `>` or `==`.** `[[ 10 > 9 ]]` compares text and is false. Use `-gt`, `-eq`.
- **No `set -euo pipefail`.** The script reports success after half of it failed.
- **Forgetting that `grep` "fails" when nothing matches.** Under `set -e`, a clean log stops the script. Add `|| true` where no matches is fine.
- **Errors on stdout.** They end up in the output file. Use `>&2`.

## Check yourself

1. In bash, `file="release notes.txt"` is followed by `rm $file`. How many arguments does rm receive, and how do you fix it?
2. In bash, what is the difference between `echo "$USER"` and `echo '$USER'`?
3. Why does the bash line `[ $env = prod ]` fail with "unary operator expected" when `env` is empty?
4. Why is the bash test `[[ 10 > 9 ]]` false, and what should you write to compare numbers?
5. What does each part of `set -euo pipefail` do in a bash script?
6. Why must `sort` come before `uniq -c` when counting repeated log messages?
7. Why does `find ... -print0 | xargs -0 rm` handle file names with spaces, when plain `find ... | xargs rm` does not?
8. A bash function needs to give back a version string like `1.4.2`. Why can't it use `return`, and what does it do instead?
9. In bash, what exit status means success, and when does the second command run in `a && b`? In `a || b`?
10. A bash script with `set -euo pipefail` runs `count="$(grep -c ERROR app.log)"`. Why does it stop when the log has no errors, and how do you allow zero matches?
11. A bash script runs `export DEPLOY_ENV=prod` and then `./check.sh`. Another sets `deploy_env=prod` without `export` and runs `./check.sh`. Which `check.sh` can read the value, and why?
12. Why does the bash line `sort app.log > app.log` leave `app.log` empty?

## Answers

1. Two: the unquoted `$file` is split at the space into `release` and `notes.txt`. Write `rm "$file"`; double quotes keep the value as one argument.
2. Double quotes expand variables, so `"$USER"` prints your user name. Single quotes keep everything literal, so `'$USER'` prints the text `$USER`.
3. Unquoted, the empty variable disappears from the line, so `[` sees only `= prod`, which is not a valid test. `[ "$env" = prod ]` passes an empty string and works.
4. `>` inside `[[ ]]` compares strings character by character, and `1` sorts before `9`. Use the numeric operators: `[[ 10 -gt 9 ]]`.
5. `-e` exits when a command fails; `-u` treats unset variables as errors, catching typos; `-o pipefail` makes a pipeline fail when any stage fails, not only the last one.
6. `uniq` only merges duplicate lines that are next to each other. `sort` puts equal lines together first, so each message is counted once in total.
7. `-print0` separates names with a zero byte and `xargs -0` splits only on that byte. Names can contain spaces and newlines but never a zero byte; plain `xargs` splits on spaces and newlines.
8. `return` only sets the exit status, a number from 0 to 255. The function prints the result with `echo`, and the caller captures it: `version="$(get_version)"`.
9. 0 means success; 1–255 mean failure. In `a && b`, `b` runs only if `a` succeeded (exited 0). In `a || b`, `b` runs only if `a` failed, as in `cd /srv/app || exit 1`.
10. `grep` exits 1 when it finds no matching lines, and `set -e` treats any non-zero status as a failure. Write `count="$(grep -c ERROR app.log || true)"`: grep still prints `0`, and `|| true` makes the line succeed.
11. Only the first. `export` puts the variable into the environment, which every child process gets a copy of. A plain variable stays inside the shell that set it.
12. The shell sets up `> app.log` before `sort` starts, and `>` empties the file. By the time `sort` reads it, there is nothing left. Write to a new file and `mv` it over the old one.

## Key terms

| Term | Meaning |
|---|---|
| shell | a program that reads command lines and starts other programs |
| terminal | the window that shows text and sends keystrokes to the shell |
| bash | the Bourne Again SHell, the default shell on most Linux servers |
| argument | one word passed to a program after the command name |
| option (flag) | an argument starting with `-` or `--` that changes behaviour |
| word splitting | the shell splitting an unquoted expansion's result at spaces |
| globbing | the shell expanding `*`, `?` and `[ ]` into matching file names |
| double quotes | expand `$` but prevent word splitting and globbing |
| single quotes | keep everything inside literally, no expansion |
| command substitution | `$(cmd)`: replaced by the command's output |
| `"$@"` | all of a script's or function's arguments, each kept as one word |
| exit status | a command's result number in `$?`; 0 means success |
| `[[ ]]` | bash's test syntax for files, strings and numbers; answers with exit status 0 (true) or 1 (false) |
| shebang | the `#!` first line naming the program that runs a script |
| `set -euo pipefail` | bash options that stop a script on errors, unset variables and failed pipeline stages |
| pipeline | commands joined by `\|`, each one's output feeding the next |
| absolute / relative path | a path from the root `/` / a path from the current folder |
| hidden file | a file whose name starts with `.`; `ls -a` shows it |
| environment | the `NAME=value` variables a process gets a copy of; `export` adds a shell variable to it |
| here-doc | `<<'EOF'` ... `EOF`: lines written in the script, fed to a command's stdin |

## Watch

- [Lecture 1: Course Overview + The Shell (2020)](https://www.youtube.com/watch?v=Z56Jmr9Z34Q) · Missing Semester · 48 min: MIT's gentle start: the prompt, commands, arguments, paths and connecting programs.
- [Lecture 2: Shell Tools and Scripting (2020)](https://www.youtube.com/watch?v=kgII-YWo3Zw) · Missing Semester · 49 min: variables, quoting, `$?`, `&&`, globs, functions and `find`, live in a terminal; the first half matches this guide.
- [Lecture 4: Data Wrangling (2020)](https://www.youtube.com/watch?v=sz_dsktIjt4) · Missing Semester · 50 min: grep, sed, sort, uniq and awk chained into pipelines on a real server log.
- [Where GREP Came From - Computerphile](https://www.youtube.com/watch?v=NTfOnGZUZDk) · Computerphile · 10 min: Brian Kernighan tells how grep was written overnight, and why small tools that do one thing work so well.

## Go deeper

- [Shell Tools and Scripting](https://missing.csail.mit.edu/2020/shell-tools/): the lecture notes and exercises for the MIT lecture above, good for practice away from this course.
- [Quotes (Greg's Wiki)](https://mywiki.wooledge.org/Quotes): the clearest full explanation of bash quoting; read it whenever quotes confuse you.
- [Bash Pitfalls](https://mywiki.wooledge.org/BashPitfalls): a numbered list of common bash bugs, each with the fix; most are quoting.
- [ShellCheck](https://www.shellcheck.net/): paste a script and get its bugs explained; also runs in your editor and in CI.
- [Google Shell Style Guide](https://google.github.io/styleguide/shellguide.html): how a large company writes bash, including when to switch to a real language.
