# Sort and uniq: top-N reports

"Which service throws the most errors?" "Who is hammering the API?" Ops
questions like these are answered with a pipe that ends in `sort`, `uniq`
and `head`. You'll build four such reports.

**New here:** `sort`, `sort -n`, `sort -r`, `sort -u`, `uniq`, `uniq -c`,
`head -n`, and `cut -d' '` to pick a column.

## Learn

### sort: put lines in order

`sort` reads lines (from a file, or from a pipe) and prints them in order.
Say `regions.txt` holds:

```text
us-east-1
eu-west-1
us-east-1
ap-south-1
eu-west-1
us-east-1
```

```bash
sort regions.txt
# ap-south-1
# eu-west-1
# eu-west-1
# us-east-1
# us-east-1
# us-east-1
```

`sort -u` (*unique*) prints each different line once:

```bash
sort -u regions.txt
# ap-south-1
# eu-west-1
# us-east-1
```

### Numbers: sort -n, and backwards: sort -r

Plain `sort` compares text character by character, like a dictionary. For
numbers that's wrong: `"9"` comes after `"10"`, because `9` is after `1`.
`-n` compares lines by the **number** they start with. `-r` *reverses* the
order, so `-rn` means "biggest number first".

```bash
printf '9 retries\n10 retries\n200 retries\n' | sort
# 10 retries
# 200 retries
# 9 retries
printf '9 retries\n10 retries\n200 retries\n' | sort -rn
# 200 retries
# 10 retries
# 9 retries
```

| Command | Order |
|---|---|
| `sort` | text, A to Z |
| `sort -r` | text, Z to A |
| `sort -n` | numbers, smallest first |
| `sort -rn` | numbers, biggest first |

> **Why is it written like this?** **sort requires** nothing extra for text;
> numbers need `-n`, because text order is the default. **Convention:**
> single-letter flags can be glued together: `-rn` is the same as `-r -n`.

### uniq: squeeze repeated lines, and count them

`uniq` drops a line when it's the same as the line **right above it**. So it
only finds every duplicate when equal lines sit next to each other, which is
exactly what `sort` does. That's why you nearly always see `sort | uniq`.

`uniq -c` *counts* each run of equal lines and puts the count in front:

```bash
sort regions.txt | uniq -c
#       1 ap-south-1
#       2 eu-west-1
#       3 us-east-1
```

Without the `sort`, `uniq` sees `us-east-1` three times in three separate
places and counts three runs of 1.

The counts are padded with spaces on the left; how many spaces depends on
the system. Add `sort -rn` to put the biggest count first, and you have the
classic "most common lines" pipe:

```bash
sort regions.txt | uniq -c | sort -rn
#       3 us-east-1
#       2 eu-west-1
#       1 ap-south-1
```

> **Why is it written like this?** **uniq requires** sorted input to catch
> every duplicate: it only ever compares a line with the one before it, so
> it can handle files bigger than memory. **Convention:** `sort | uniq -c |
> sort -rn` is a recipe you'll type for the rest of your career.

### head -n: keep the first lines

`head -n 2` prints only the first 2 lines it receives. At the end of a
sorted pipe, that's a *top-N*:

```bash
sort regions.txt | uniq -c | sort -rn | head -n 2
#       3 us-east-1
#       2 eu-west-1
```

N can come from a variable, with a default as in *Shell variables*:
`head -n "${limit:-10}"`.

### cut -d' ': pick one column

Log lines are columns separated by spaces. `cut` (you met it in *Grep and
pipes*) splits each line at a delimiter and keeps the fields you ask for.
With `-d' '`, the delimiter is a space (quoted, or bash would swallow it):

```bash
echo 'node-3 Ready 41d v1.31.2' | cut -d' ' -f2
# Ready
```

> **Why is it written like this?** **cut requires** the delimiter to be one
> character, so two spaces in a row count as an empty field between them.
> The files in this exercise have exactly one space between columns.
> *Columns with cut and awk* shows the tool for messier columns.

## Your task

Build the reports the on-call engineer asks for every morning. Each script
takes a file as `$1`. The fixtures are in `testdata/`: open them and check
the answers by eye.

- `testdata/app.log` and `testdata/worker.log`: app logs. Every line is
  `TIME LEVEL SERVICE msg="..." req=ID`, with one space between the columns,
  for example:

  ```text
  2026-10-13T09:15:09Z ERROR payments msg="card declined" req=651327
  ```

- `testdata/du_web01.txt` and `testdata/du_db01.txt`: disk usage, as
  `du -sk` prints it: a size in kilobytes, a tab, a directory.
- `testdata/access.log` and `testdata/access_2.log`: nginx access logs. The
  first column is the client IP, followed by a space.

### 1. List the services

**Goal:** make `1_services.sh` print every service name (the third column)
that appears in the log `$1`, each once, sorted A to Z.

```bash
bash 1_services.sh testdata/worker.log
# exporter
# mailer
# scheduler
# worker
```

The tests also run it on `testdata/app.log` (5 services).

**Think it through:**
- Which command keeps only the third column, and what is the delimiter?
- Which flag makes `sort` print each line once?

### 2. Find the biggest directories

**Goal:** make `2_biggest_dirs.sh` print the 5 biggest lines of the disk
usage file `$1`, biggest first, each line unchanged.

```bash
bash 2_biggest_dirs.sh testdata/du_web01.txt
# 30412	/opt/app/releases
# 12804	/var/lib/docker
# 2310	/var/log/journal
# 1024	/var/log/app
# 880	/var/cache/apt
```

The tests also run it on `testdata/du_db01.txt`.

**Think it through:**
- Do you want text order or number order? Try plain `sort` first and look.
- Which flag puts the biggest first, and which command keeps the first 5?

### 3. Count errors per service

**Goal:** make `3_errors_per_service.sh` print, for each service with
`ERROR` lines in the log `$1`, the number of `ERROR` lines and the service
name, highest count first.

```bash
bash 3_errors_per_service.sh testdata/app.log
#       4 payments
#       2 api
#       1 checkout
```

Use `uniq -c` for the counts. The tests **ignore the spaces at the start of
each line**, so `4 payments` and `      4 payments` both pass. Only `ERROR`
lines count, not `WARN` or `INFO`. The tests also run it on
`testdata/worker.log` (3 mailer, 2 worker, 1 scheduler).

**Think it through:**
- Build it step by step: the ERROR lines, then the service column, then...?
- What must happen before `uniq -c` can count each service in one go?

### 4. Find the top client IPs

**Goal:** make `4_top_ips.sh` print the N client IPs (first column) with the
most requests in the access log `$1`, each with its count, most first. N is
`$2`, or `3` when there's no `$2`.

```bash
bash 4_top_ips.sh testdata/access.log
#       8 203.0.113.7
#       6 198.51.100.23
#       5 192.0.2.44
bash 4_top_ips.sh testdata/access.log 1
#       8 203.0.113.7
```

As in task 3, the tests ignore spaces at the start of each line. They also
run `testdata/access.log 5` and `testdata/access_2.log` (no `$2`).

**Think it through:**
- Which column holds the IP, and how do you keep only it?
- How do you count, order and cut the list down to N lines?
- How do you write "`$2`, or 3 if it's missing"?

## Learn more

- [GNU coreutils: sort invocation](https://www.gnu.org/software/coreutils/manual/html_node/sort-invocation.html)
- [GNU coreutils: uniq invocation](https://www.gnu.org/software/coreutils/manual/html_node/uniq-invocation.html)
- [The Missing Semester: Data wrangling](https://missing.csail.mit.edu/2020/data-wrangling/)
