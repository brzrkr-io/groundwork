# Columns with cut and awk

Inventories, CSV exports and access logs are all tables: rows of fields.
`cut` grabs columns by number, and `awk` can also filter rows and add up
numbers, all in one line you can type on any server.

**New here:** `cut -d, -f`, `tail -n +2`, `awk` with fields (`$1`, `$NF`),
`-F,`, patterns like `$3 == "x"` and `$5 > 100`, sums with `END`.

## Learn

### cut: keep some columns

`cut -d CHAR -f LIST` splits every line at CHAR (the *delimiter*) and prints
the fields in LIST, counted from 1. Say `services.csv` holds:

```text
service,port,team
api,8080,core
billing,9090,payments
```

```bash
cut -d, -f1 services.csv
# service
# api
# billing
cut -d, -f1,3 services.csv
# service,core
# api,core
# billing,payments
```

With several fields, cut puts the delimiter back between them. `-f2-`
means "field 2 to the end".

> **Why is it written like this?** **cut requires** a one-character
> delimiter. `-d,` needs no quotes because a comma means nothing to bash;
> a space does, so a space delimiter is written `-d' '`.

### tail -n +2: skip the header

`head -n 3` keeps the first 3 lines. Its partner `tail` keeps the **last**
lines, and with a `+` it means "start at line N": `tail -n +2` prints
everything from line 2 on, which drops a CSV header.

```bash
tail -n +2 services.csv
# api,8080,core
# billing,9090,payments
```

### awk: a tiny language for columns

`awk` runs a small program on every line. The program goes in single
quotes. Inside `{ }` is what to do; `$1`, `$2`, ... are the line's fields,
`$0` the whole line, and `$NF` the last field (`NF` is the *number of
fields*).

```bash
echo 'node-3   Ready   41d   v1.31.2' | awk '{print $2, $NF}'
# Ready v1.31.2
```

By default awk splits on **runs** of spaces and tabs, so ragged columns are
fine (that's where `cut -d' '` gives up). A comma in `print` puts a single
space in the output; text in double quotes is printed as it is.

For a CSV, set the *field separator* with `-F,`:

```bash
awk -F, '{print $1 " on port " $2}' services.csv
# service on port port
# api on port 8080
# billing on port 9090
```

Two things side by side, like `$1 " on port "`, are joined together.

> **Why is it written like this?** **awk requires** its program as one
> argument, hence the single quotes: they also stop bash from swapping in
> `$1` itself. **Convention:** `print a, b` for space-separated output,
> `print a "text" b` when you need exact text in between.

### Patterns: only some lines

Put a condition before the `{ }` and awk runs the action only on lines where
it's true. Compare text with `==` (text in double quotes) and numbers with
`<`, `<=`, `>`, `>=`:

```bash
awk -F, '$3 == "core" {print $1}' services.csv
# api
tail -n +2 services.csv | awk -F, '$2 > 9000 {print $1}'
# billing
```

`==` checks the **whole** field: `"core"` doesn't match `core-infra`, unlike
`grep core`. The number example skips the header first: compared with a
number, a word like `port` is compared as text, and the answer surprises.

### Totals: variables and END

awk variables need no `$` and start at 0. `n += $2` adds field 2 to `n` on
every line. A block labelled `END` runs once, after the last line, which is
the moment to print the total:

```bash
awk -F, 'NR > 1 {count += 1} END {print "services: " count}' services.csv
# services: 2
```

(`NR` is the line number, so `NR > 1` skips the header.)

> **Why is it written like this?** **awk requires** `END` to be spelled in
> capitals. **Convention:** short lowercase variable names. **Your choice:**
> one awk program like this, or a pipe of smaller tools: both are fine when
> they're easy to read.

## Your task

You're on call with an inventory and two access logs. Open the fixtures
first:

- `testdata/hosts.csv` and `testdata/hosts_eu.csv`: host inventories with
  the header `name,ip,env,role`.
- `testdata/access.log` and `testdata/access_2.log`: nginx access logs in
  the standard *combined* format, one request per line:

  ```text
  192.0.2.44 - - [13/Oct/2026:09:00:14 +0000] "POST /api/checkout HTTP/1.1" 502 157 "-" "curl/8.5.0"
  ```

  That's the client IP, two dashes, the time, the request (method, path,
  protocol), the status code, the response size in bytes, the referrer and
  the user agent.

Each script takes the file as `$1`.

### 1. List names and IPs

**Goal:** make `1_names_and_ips.sh` print the name and IP of every host in
the CSV `$1`, as `name,ip`, without the header line.

```bash
bash 1_names_and_ips.sh testdata/hosts_eu.csv
# eu-web-01,10.20.1.11
# eu-web-02,10.20.2.12
# eu-db-01,10.20.1.21
# eu-queue-01,10.20.3.31
# eu-db-02,10.20.1.22
```

The tests also run it on `testdata/hosts.csv` (9 hosts).

**Think it through:**
- Which fields are the name and the IP, and what separates them?
- How do you drop the first line before (or after) cutting?

### 2. List the production hosts

**Goal:** make `2_prod_hosts.sh` print the name of every host in the CSV
`$1` whose env is exactly `prod`, one per line, in file order.

```bash
bash 2_prod_hosts.sh testdata/hosts_eu.csv
# eu-web-01
# eu-db-01
# eu-db-02
```

`preprod` is not `prod`. The tests also run it on `testdata/hosts.csv`
(5 production hosts).

**Think it through:**
- Which field is the env, and how do you tell awk the fields are split by
  commas?
- What condition picks the line, and what do you print?

### 3. Find the server errors

**Goal:** make `3_server_errors.sh` print `STATUS PATH` (one space between)
for every request in the access log `$1` with a status of 500 or more, in
file order.

```bash
bash 3_server_errors.sh testdata/access.log
# 502 /api/checkout
# 500 /api/orders
# 503 /api/search?q=boots
# 504 /api/orders
```

4xx statuses like 404 are the client's fault and stay out. The tests also
run it on `testdata/access_2.log`.

**Think it through:**
- Count the fields of the sample line (awk splits on spaces). Which number
  is the status, and which the path? Check with `awk '{print $7}'`.
- Which comparison means "500 or more"?

### 4. Add up the bytes

**Goal:** make `4_total_bytes.sh` print `total bytes: N`, where N is the sum
of the response sizes of every request in the access log `$1`.

```bash
bash 4_total_bytes.sh testdata/access.log
# total bytes: 65684
bash 4_total_bytes.sh testdata/access_2.log
# total bytes: 18218
```

**Think it through:**
- Which field is the response size? (It's right after the status.)
- What runs on every line, and what runs once at the end?

## Learn more

- [GNU coreutils: cut invocation](https://www.gnu.org/software/coreutils/manual/html_node/cut-invocation.html)
- [GNU awk: Getting started](https://www.gnu.org/software/gawk/manual/html_node/Getting-Started.html)
- [The Missing Semester: Data wrangling](https://missing.csail.mit.edu/2020/data-wrangling/)
