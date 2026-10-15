# Editing text with sed

Bumping an image tag, hiding secrets before you paste a config into a ticket,
flipping a log level on a server: all of these are "change some lines in a
text file". `sed` does that from the command line and from scripts.

**New here:** `sed 's/old/new/'`, the `g` flag, `sed -E` and groups (`\1`),
other delimiters (`s|a|b|`), `/pattern/d`, `sed -n` with `p`, `sed -e`,
`sed -i`, a variable inside a sed script.

## Learn

### sed edits a stream of lines

`sed` (the *stream editor*) reads a file line by line, runs a small *script*
on each line, and prints the result. The file itself stays as it was:

```bash
sed 's/staging/production/' hosts.txt
```

This prints `hosts.txt` with the first `staging` on each line changed to
`production`. Lines without `staging` are printed unchanged. With no file
name, sed reads from a pipe instead: `cat hosts.txt | sed '...'`.

> **Why is it written like this?** **sed requires** the script as one
> argument, so it goes in quotes. **Convention:** single quotes, so bash
> leaves every character for sed.

### s/old/new/: substitute

`s` is sed's *substitute* command. Between the slashes come the pattern to
find and the text to put in its place. It changes the **first** match on each
line. Add `g` (*global*) at the end to change every match:

```bash
echo 'retry retry retry' | sed 's/retry/try/'
# try retry retry
echo 'retry retry retry' | sed 's/retry/try/g'
# try try try
```

### Patterns

The pattern is a *regular expression*: plain text, plus a few characters
with a special meaning.

| Pattern | Matches |
|---|---|
| `^` | the start of the line |
| `$` | the end of the line |
| `.` | any one character |
| `.*` | any run of characters, as long as possible (also none) |
| `[0-9]` | one digit |
| `\.` | a real dot |

```bash
echo 'timeout=30s' | sed 's/=.*/=60s/'
# timeout=60s
```

`=.*` means "an `=` and everything after it, to the end of the line". Watch
out: a pattern matches **anywhere** in the line unless you pin it with `^`
or `$`. `s/out=/.../` would also match inside `timeout=`.

A `.` matches any character, so `1.4` also matches `104`. That rarely gets in
the way; write `1\.4` when you need to be exact.

### -E and groups: keep part of the match

With `-E` (*extended* regular expressions), a few more characters get a
special meaning:

- `+` means "one or more of the thing before it": `[0-9]+` is a number.
- `(FAILED|ERROR)` means `FAILED` or `ERROR`.
- Parentheses `( )` make a *group*: sed remembers what that part matched.
- In the new text, `\1` is what group 1 matched, `\2` group 2, and so on.

```bash
echo 'status: FAILED' | sed -E 's/(FAILED|ERROR)/[\1]/'
# status: [FAILED]
echo '2026-10-13' | sed -E 's/([0-9]+)-([0-9]+)-([0-9]+)/\3.\2.\1/'
# 13.10.2026
```

> **Why is it written like this?** **sed requires** `-E` for `( )` and `|`
> to have these meanings; without it they're plain characters. **Convention:**
> always use `-E` when you need groups. It works the same in GNU sed (Linux)
> and BSD sed (macOS).

### Another delimiter for paths

The character right after `s` is the *delimiter*. A `/` is usual, but it's
awkward when the text itself has slashes, like paths and URLs. Any character
works; `|` is popular:

```bash
echo 'log_dir=/var/log/app' | sed 's|/var/log|/data/logs|'
# log_dir=/data/logs/app
```

> **Why is it written like this?** **sed requires** the same delimiter three
> times. **Your choice:** which one. With `/` you'd have to write every slash
> in the text as `\/`.

### Deleting and printing lines

Put a pattern between slashes **before** a command, and the command runs only
on lines that match. `d` deletes the line (it isn't printed):

```bash
sed '/^DEBUG/d' app.log        # every line except those starting with DEBUG
```

`-n` turns off the automatic printing, and `p` prints a line on purpose.
Together they print only the matching lines, like `grep`:

```bash
sed -n '/timeout/p' app.log    # only the lines that contain timeout
```

To run several commands, give each one its own `-e`. sed runs them in order
on every line:

```bash
sed -e '/^DEBUG/d' -e 's/WARN/warning/' app.log
```

### A variable in the script: double quotes

Inside single quotes, bash leaves `$env` alone, so sed would look for the
text `$env`. To use a variable's value (or `$1`), put the script in **double
quotes**, as in *Shell commands and quoting*:

```bash
env=production
sed "s/ENVIRONMENT/$env/" template.txt
```

> **Why is it written like this?** **Bash requires** double quotes for `$`
> to swap in a value. **Convention:** single quotes when the script has no
> variables, so a `$` meaning "end of line" stays with sed.

### -i: edit the file in place

Normally sed only prints. `-i` (*in place*) writes the result back into the
file instead, and prints nothing:

```bash
sed -i 's/replicas: 2/replicas: 4/' deploy.yaml
```

There is no undo. Run the command **without** `-i` first, check the output,
then add `-i`. Never point `-i` at the files in `testdata/`: the tasks read
them. (The tests run your scripts on a copy, and tell you if one changed.)

> **Why is it written like this?** This is GNU sed, the one on Linux servers
> and in the tests. On macOS, `sed -i` needs an empty argument after it:
> `sed -i '' 's/.../.../' file`. If you run the scripts on a Mac, that's the
> one difference.

### When sed complains

sed's error names the place: `-e expression #1, char 14` means "your
first script, at character 14". Count along the script to find the spot.

| sed says | Usually means |
|---|---|
| `unknown option to 's'` | the text has the delimiter in it, like the `/` in a path: pick another delimiter |
| `unterminated 's' command` | a delimiter is missing: `s/a/b` needs a third `/` |
| `Unmatched ( or \(` | with `-E`, a `(` without its `)` |
| `can't read FILE` | the file name is wrong, or you're not in the exercise folder |

If sed prints the file unchanged, the pattern never matched. Try a shorter
piece of it, then grow it back.

## Your task

You're on the platform team. Today: a tag bump for the `api` deployment, and
two env files that need cleaning up before you share them. Look at
`testdata/deploy.yaml`, `testdata/app.env` and `testdata/worker.env` first.

Each task is one script, and each script needs one `sed` command. Run them
from the exercise folder, as in the examples.

### 1. Bump the api image tag

**Goal:** make `1_set_image_tag.sh` print `testdata/deploy.yaml` with the
`api` container's image tag replaced by the tag given as `$1`.

Only this line changes; it keeps its indentation:

```text
          image: registry.example.com/api:1.4.1
```

Every other line is printed exactly as it is in the file. That includes the
`log-shipper` image, which also has the tag `1.4.1` and must keep it.

```bash
bash 1_set_image_tag.sh 1.5.0
# ... (every line before it, unchanged)
#         - name: api
#           image: registry.example.com/api:1.5.0
# ...
#         - name: log-shipper
#           image: registry.example.com/log-shipper:1.4.1
# ... (every line after it, unchanged)
```

The tests run it with `1.5.0` and with `2.0.0-rc.1`.

**Think it through:**
- Which text appears **only** on the line you want to change?
- The image name has slashes. Which delimiter saves you trouble?
- How does `$1` get into the sed script?

### 2. Redact the secrets

**Goal:** make `2_redact_secrets.sh` print the env file given as `$1` with
every secret value replaced by `REDACTED`.

A *secret* is a line whose key (the part before the `=`) ends in `_PASSWORD`
or `_TOKEN`. Print its key, then `=`, then `REDACTED`. Print every other line
unchanged, comments and empty lines included.

```bash
bash 2_redact_secrets.sh testdata/app.env
```

prints:

```text
# api service settings
# format: KEY=value, one per line
# set LOG_LEVEL=debug only while troubleshooting

APP_NAME=api
APP_PORT=8080
LOG_LEVEL=info
ALERT_CHANNEL=#api-alerts

# database
DB_HOST=db.internal
DB_USER=api
DB_PASSWORD=REDACTED
PASSWORD_MIN_LENGTH=12

# third-party services
STRIPE_API_TOKEN=REDACTED
SLACK_TOKEN=REDACTED
TOKEN_TTL_SECONDS=3600
```

The tests also run it on `testdata/worker.env`, where `QUEUE_PASSWORD` and
`METRICS_TOKEN` are the secrets.

**Think it through:**
- `PASSWORD_MIN_LENGTH` and `TOKEN_TTL_SECONDS` are not secrets. What comes
  right after `_PASSWORD` or `_TOKEN` in a key that ends with it?
- Which part of the line do you replace, and which part must survive?
- A value can contain `=` (look at `METRICS_TOKEN`). Is all of it gone?

### 3. Strip comments and empty lines

**Goal:** make `3_strip_comments.sh` print the env file given as `$1`
without its comment lines and without its empty lines.

A *comment line* starts with `#`. An *empty line* has nothing on it. Every
other line is printed unchanged, in order.

```bash
bash 3_strip_comments.sh testdata/worker.env
# WORKER_CONCURRENCY=8
# LOG_LEVEL=info
# QUEUE_URL=amqp://mq.internal:5672/jobs
# QUEUE_PASSWORD=rabbit-Ears-77
# METRICS_TOKEN=bXQtOWY4ZTdk==
# METRICS_PORT=9100
```

The tests also run it on `testdata/app.env`, which has 11 settings left
afterwards.

**Think it through:**
- `ALERT_CHANNEL=#api-alerts` has a `#` too, and must stay. Where exactly is
  the `#` in a comment line?
- How do you write "a line with nothing between its start and its end"?
- Two kinds of lines to drop: one command or two?

### 4. Change the log level in place

**Goal:** make `4_set_log_level.sh` edit the file given as `$1` so that its
`LOG_LEVEL` line has the value given as `$2`.

- Change only the line that **starts with** `LOG_LEVEL=`. The comment
  `# set LOG_LEVEL=debug only while troubleshooting` stays as it is.
- Every other line stays exactly as it was.
- The script prints nothing; it changes the file.

```bash
cp testdata/app.env copy.env
bash 4_set_log_level.sh copy.env debug
grep '^LOG_LEVEL' copy.env
# LOG_LEVEL=debug
```

Work on a copy like `copy.env`, so `testdata/` stays as it is. The tests do
the same: they copy `testdata/app.env` and `testdata/worker.env` to
`copy.env` in a temporary folder, set the levels `debug` and `warn`, and
compare the whole file afterwards.

**Think it through:**
- Which flag writes the result back into the file?
- How do you make the pattern match only at the start of a line?
- This time there are two variables: which is the file, and which goes into
  the script?

## Learn more

- [GNU sed manual](https://www.gnu.org/software/sed/manual/sed.html)
- [The Missing Semester: Data Wrangling](https://missing.csail.mit.edu/2020/data-wrangling/)
