# Shell commands and quoting

*Practice: reproduce a bug report* closed *Testing*, and with it the Go
foundations. Now you step outside Go to the shell that builds, runs and
glues together every program you write.

Every server, container and CI job has a shell. You type commands into it, and
scripts are just those commands saved in a file. This first module outside Go
starts with finding your way around, then the two things every script does:
print text, and quote it right.

**New here:** the terminal, `pwd`, `ls`, `cd`, `cat`, running a script with
`bash`, `echo`, `printf`, single and double quotes.

## Learn

### Terminal, shell, command

| Word | What it is |
|---|---|
| *terminal* | the window you type into. Groundwork is running in one. |
| *shell* | the program inside it that reads your line and runs it. Here: *bash*. |
| *command* | a program the shell runs, like `ls` or `echo` |

Open a **second** terminal for this module: a new tab or window, or your
editor's terminal panel (VS Code: **Terminal → New Terminal**). Groundwork
keeps checking your work in the first one.

The text before your cursor (often ending in `$`) is the *prompt*. Type a
command, press **Enter**, and its output appears below. **↑** brings back
your last command.

In the examples here, everything after a `#` is a note for you, not
something to type. A `#` line under a command shows what it prints.

### Get to the exercise folder

| Command | What it does |
|---|---|
| `pwd` | prints the folder you're in |
| `ls` | lists the files in it |
| `cd FOLDER` | moves into FOLDER; `cd ..` moves up one |
| `cat FILE` | prints a file |

Every exercise has its own folder inside your course folder: `exercises/`,
then the module, then the exercise, for example
`exercises/08_shell/145_shell_commands/`. Groundwork shows the file to work
on under **Edit** (or on its `edit` line in a narrow window); the folder is
that path without the file name.

To get a terminal there, right-click the exercise folder in your editor's
file list and choose **Open in Integrated Terminal** (VS Code). Or `cd`
there yourself: `pwd` shows where you are, `ls` what's around you, and you
go down one folder at a time.

```bash
pwd
# /home/sam/groundwork/course        (your course folder; yours will differ)
cd exercises/08_shell/145_shell_commands   # use the folder Groundwork shows
ls
# 1_deploy_line.sh  2_reminder.sh  3_status_table.sh  README.md ...
```

**Tab** finishes a name for you: type `cd exer` and press Tab, and bash
fills in `exercises/`. Press Tab twice to list the choices. Use it all the
time; it saves typing and typos.

### A script is a list of commands

A *script* is a text file of commands that bash runs from top to bottom.
This exercise has one script per task: `1_deploy_line.sh`, `2_reminder.sh`
and `3_status_table.sh`. Edit one in your editor and save. Groundwork re-runs
the tests straight away. To see the output yourself, run it from the
exercise folder:

```bash
bash 1_deploy_line.sh
```

The tests run each script exactly like that and compare what it prints with
what the task asks for.

### The first two lines

Every script here starts the same way:

```bash
#!/usr/bin/env bash
set -euo pipefail
```

| Line | Meaning |
|---|---|
| `#!/usr/bin/env bash` | the *shebang*: "run this file with bash" |
| `set -euo pipefail` | stop at the first failing command instead of carrying on |

Leave both lines alone. Everything else starting with `#` is a *comment*:
bash ignores it.

> **Why is it written like this?** **Bash requires** nothing here: a script
> runs without them. **Convention:** every serious script has the shebang so
> it can run as `./script.sh`, and `set -euo pipefail` so a mistake stops it
> early, before it does damage. The *Shell variables* exercise explains each
> letter.

### Commands and arguments

A command line is a command name followed by *arguments*, separated by
spaces. Type these at your prompt and watch what happens:

```bash
echo deploy finished
# deploy finished
```

Here `echo` is the command, and `deploy` and `finished` are two arguments.
`echo` prints its arguments with one space between them, then a newline.

Bash splits the line into words **before** `echo` sees it. Extra spaces
vanish, and some characters, like `(`, `)`, `;`, `|`, `$`, `*` and `"`, mean
something to bash itself:

```bash
echo deploy    finished
# deploy finished        (the extra spaces are gone)

echo build (debug)
# bash: syntax error near unexpected token `('
```

### Quotes keep text in one piece

*Quotes* tell bash "this is one argument, leave it alone". Bash removes the
quotes before running the command, so `echo` never sees them.

```bash
echo "build (debug)    done"
# build (debug)    done
```

There are two kinds, and the difference is the `$` sign. A `$` asks bash to
swap in a variable's value (*Shell variables* is all about that). `$HOME`,
for example, holds your home folder:

| Quotes | Inside them | Example | Prints |
|---|---|---|---|
| `"double"` | `$` still swaps in values | `echo "home: $HOME"` | `home: /home/sam` |
| `'single'` | **everything** is plain text | `echo 'home: $HOME'` | `home: $HOME` |

A double quote inside single quotes is just a character, and so is a single
quote inside double quotes:

```bash
echo 'She said "ship it"'
# She said "ship it"
echo "It's live"
# It's live
```

> **Why is it written like this?** **Bash requires** quotes to come in pairs
> of the same kind. **Convention:** use single quotes for fixed text, and
> double quotes when you want a variable's value inside. When unsure, quote:
> quotes around plain text never hurt.

### printf: formatted output

`printf` works like Go's `fmt.Printf`. The first argument is the *format*,
and each `%s` takes the next argument. Unlike `echo`, it adds no newline: you
write `\n` yourself.

```bash
printf '%s has %s pods\n' api 3
# api has 3 pods
```

A width goes between `%` and `s`. `%-8s` pads the value with spaces to 8
characters, on the right (the `-` means "line up on the left"):

```bash
printf '[%-8s]\n' web
# [web     ]
```

If you give more arguments than the format uses, `printf` repeats the format
for the rest. One call can print a whole table.

> **Why is it written like this?** **Convention:** `printf` for anything with
> columns or exact formatting, `echo` for a simple line. The format goes in
> single quotes so bash leaves `%` and `\n` for `printf` to handle.

## Your task

You're writing the messages a deploy script prints. Every space, quote and
`$` counts.

### 1. Print the deploy line

**Goal:** make `1_deploy_line.sh` print exactly `Deploying checkout (3 replicas)`.

```bash
bash 1_deploy_line.sh
# Deploying checkout (3 replicas)
```

Try it without quotes first and read what bash says.

**Think it through:**
- Which characters in that line mean something to bash?
- How do you tell bash to treat the whole line as plain text?

### 2. Print the reminder

**Goal:** make `2_reminder.sh` print exactly these two lines:

```text
Set $IMAGE_TAG before you deploy.
Then run "make deploy".
```

The `$IMAGE_TAG` and the double quotes must appear as written. Use one `echo`
per line.

**Think it through:**
- In double quotes, what would bash do with `$IMAGE_TAG`? (With `set -u`, an
  unknown variable stops the script.)
- Which kind of quotes leaves `$` alone, and can also hold a `"`?

### 3. Print the status table

**Goal:** make `3_status_table.sh` print this table using `printf`:

```text
SERVICE    STATUS
api        running
worker     stopped
payments   running
```

The first column is padded to 10 characters, then comes one space, then the
second column, then a newline. The header is a row like the others.

**Think it through:**
- What format pads a value to 10 characters, lined up on the left?
- What separates the two columns, and what ends each line?

## Learn more

- [Bash manual: Quoting](https://www.gnu.org/software/bash/manual/html_node/Quoting.html)
- [GNU coreutils: printf invocation](https://www.gnu.org/software/coreutils/manual/html_node/printf-invocation.html)
- [Greg's Wiki: Quotes](https://mywiki.wooledge.org/Quotes)
