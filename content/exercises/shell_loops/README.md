# Shell loops

Ops work is the same step on many things: every host, every line of an
inventory, every log file. Bash has two loops for that: `for` walks a list,
and `while read` walks a file line by line.

**New here:** `for ... in ...; do ... done`, `"$@"` and `$#`, arrays,
`while IFS= read -r line`, `continue`, arithmetic with `$(( ))`, globs like
`dir/*.log`, `basename`.

## Learn

### for walks a list

```bash
for region in eu-west-1 us-east-1 ap-south-1; do
	echo "deploying to $region"
done
```

The body between `do` and `done` runs once per word, with `region` set to
each one in turn.

`"$@"` is the list of all the script's arguments, each one kept whole (even
with spaces inside), and `$#` is how many there are:

```bash
# bash restart.sh api worker
for service in "$@"; do
	echo "restarting $service"
done
echo "$# services restarted"
```

> **Why is it written like this?** **Bash requires** `do` and `done` around
> the body, and `; do` when `do` is on the `for` line. **Convention:**
> always `"$@"` in double quotes; unquoted, bash would split arguments that
> contain spaces.

### Arrays

An *array* holds a list in one variable. Write it in parentheses, and read
it all back with `"${name[@]}"`:

```bash
zones=(eu-west-1a eu-west-1b eu-west-1c)
for zone in "${zones[@]}"; do
	echo "zone $zone"
done
echo "${#zones[@]} zones"     # 3 zones
```

It's like a Go slice of strings. `"$@"` is the script's own built-in array.

### while read walks a file

```bash
while IFS= read -r line; do
	echo "got: $line"
done < testdata/regions.txt
```

`< file` after `done` feeds the file into the loop. Each time round, `read`
takes one line into `line`; at the end of the file it fails, and the loop
stops.

| Piece | Why it's there |
|---|---|
| `IFS=` | keep spaces at the start and end of the line |
| `-r` | keep backslashes as they are |

Give `read` two names and it splits each line at spaces: the first word into
the first name, the rest into the second.

```bash
# each line: "web-01 10.0.0.5"
while read -r host ip; do
	echo "$host is at $ip"
done < hosts.txt
```

`continue` skips the rest of the body and goes on to the next line. A text
pattern after `==` in `[[ ]]` can end in `*`, meaning "anything":
`[[ "$line" == "//"* ]]` is true when the line starts with `//`.

### Arithmetic

Variables are text, so `total + count` would just be text. `$(( ... ))` does
whole-number maths, and inside it you can leave out the `$` on names:

```bash
pods=3
pods=$((pods * 2))
echo "$pods"      # 6
```

### Globs: every file that matches

`*` in a path is a *glob*: bash replaces it with every matching file name,
sorted alphabetically. Keep the variable part in quotes and the `*` outside:

```bash
for conf in "$dir"/*.conf; do
	echo "checking $conf"
done
```

`basename` strips the directories from a path:
`basename /etc/nginx/nginx.conf` prints `nginx.conf`.

> **Why is it written like this?** **Bash requires** the `*` to be outside
> the quotes, or it's just a star. **Convention:** loop over a glob, never
> over the output of `ls`, which breaks on names with spaces.

## Your task

Four loops from a deploy toolbox. Look in `testdata/`: two inventories, two
request-count files and two folders of logs.

### 1. Check every host

**Goal:** make `1_check_hosts.sh` print `checking HOST` for each argument, in
order, then `hosts checked: N`, where N is the number of arguments.

```bash
bash 1_check_hosts.sh web-01 web-02 db-01
# checking web-01
# checking web-02
# checking db-01
# hosts checked: 3
```

**Think it through:**
- Which list holds all the arguments?
- Which variable already knows how many there are?

### 2. Read the inventory

**Goal:** make `2_inventory.sh` print `ssh deploy@HOST uptime` for each line
of the file `$1`, skipping blank lines and lines that start with `#`.

```bash
bash 2_inventory.sh testdata/inventory_staging.txt
# ssh deploy@web-01.staging.example.com uptime
# ssh deploy@db-02.staging.example.com uptime
```

Each remaining line is one host name, used as it is.

**Think it through:**
- Which test is true for an empty line, and which for a line starting with
  `#`? (`#` must be in quotes, or bash reads a comment.)
- What skips to the next line without printing?

### 3. Total the requests

**Goal:** make `3_total_requests.sh` print `total: N`, the sum of the
numbers in the file `$1`.

```text
api 1200
web 3400
worker 85
billing 410
```

```bash
bash 3_total_requests.sh testdata/requests.txt
# total: 5095
```

Each line is a service name, a space and a count. The starter sets
`total=0` and prints it at the end; you write the loop in between.

**Think it through:**
- How do you make `read` split each line into the name and the count?
- How do you add the count to `total`, as a number?

### 4. Count lines in each log

**Goal:** make `4_log_lines.sh` print `NAME: N lines` for each file in the
directory `$1` whose name ends in `.log`, in alphabetical order.

```bash
bash 4_log_lines.sh testdata/logs
# api.log: 3 lines
# web.log: 2 lines
# worker.log: 5 lines
```

NAME is the file name without the directory. `testdata/logs/README.txt`
isn't a log, so it's left out.

**Think it through:**
- Which glob matches only the `.log` files in `$1`? Which part is quoted?
- How did *Shell variables* count lines and print only the number?

## Learn more

- [Bash manual: Looping Constructs](https://www.gnu.org/software/bash/manual/html_node/Looping-Constructs.html)
- [Greg's Wiki: BashFAQ 001](https://mywiki.wooledge.org/BashFAQ/001) (how to read a file line by line, and why `IFS=` and `-r`)
- [Bash manual: Arrays](https://www.gnu.org/software/bash/manual/html_node/Arrays.html)
