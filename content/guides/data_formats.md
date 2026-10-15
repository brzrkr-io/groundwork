# Data formats and parsing

**In one sentence:** every file and network message is just a row of bytes; a *format* is an agreement about what those bytes mean, and *parsing* is turning them into values your program can check and use.

**Why it matters:** most ops work is reading what another program wrote (nginx logs, Kubernetes manifests, `terraform.tfvars`, an AWS API response) and writing something another program will read. When a deploy breaks because of one missing quote in a YAML file, that is a data format problem.

**You'll learn:**
- why "it's all bytes" is the starting point, and what an encoding like UTF-8 does
- why you read big logs line by line instead of all at once
- how CSV, JSON, YAML, TOML and INI differ, and the YAML traps that bite in production
- the difference between parsing and validating, and why you do both
- where regular expressions shine and where they fail
- how to write timestamps that every tool and every time zone agrees on

## Everything is bytes

Open any file on a server with a hex viewer and you see the same thing: numbers from 0 to 255. A *byte* is one of those numbers. A log file, a JPEG, a compiled Go binary and a Kubernetes manifest are all just bytes in a row.

What makes them different is the agreement about what the bytes mean. That agreement is the *format*.

```go
package main

import (
	"fmt"
	"unicode/utf8"
)

func main() {
	status := "OK\n"
	fmt.Println([]byte(status))

	city := "Malmö"
	fmt.Println(len(city), utf8.RuneCountInString(city))
	fmt.Println([]byte(city))
}
```

It prints:

```text
[79 75 10]
6 5
[77 97 108 109 195 182]
```

Step by step:

1. `"OK\n"` is three bytes: 79 is `O`, 75 is `K`, and 10 is the newline character.
2. `"Malmö"` has 5 characters but 6 bytes. The `ö` takes two bytes, 195 and 182.
3. `len` counts bytes. `utf8.RuneCountInString` counts characters.

### Text and binary

**Text formats** use bytes that map to readable characters: logs, JSON, YAML, CSV. You can open them in an editor, `grep` them and diff them in Git.

**Binary formats** use bytes as raw numbers: images, `.tar.gz` archives, compiled programs, Protocol Buffers. They are smaller and faster to read by machine, but `cat` shows garbage.

Ops work leans heavily on text, because humans read it at 3 a.m. during an incident.

## Encodings: from bytes to characters

An *encoding* is the rule that maps characters to bytes.

- **ASCII** (1963) covers English letters, digits and punctuation in the numbers 0–127. One byte per character. No `ö`, no `é`, no emoji.
- **UTF-8** covers every character in Unicode. ASCII characters keep their single byte, so plain English ASCII text is already valid UTF-8. Other characters take 2 to 4 bytes.

UTF-8 is the default almost everywhere you'll work: Linux, JSON, YAML, Go source code and Go strings.

The key lesson: **there is no such thing as plain text without an encoding.** If a program reads UTF-8 bytes as if they were another encoding, `Malmö` turns into `MalmÃ¶`. When you see junk like that in a log or a web page, suspect an encoding mismatch first.

You practise this in *Bytes, runes and UTF-8*.

## Line-oriented text: logs and friends

Many ops files use the simplest format of all: one record per line, fields split by a fixed separator.

Here is one line of an nginx access log:

```text
10.0.0.9 - - [13/Oct/2026:09:15:03 +0000] "POST /deploy HTTP/1.1" 502 157
```

(Real lines also end with the referrer and the user agent, the client program's name.) The fields are separated by spaces: client IP, two fields that are almost always `-`, time, request, status code, response size.

And here is a line of `/etc/passwd`, the Linux list of user accounts:

```text
deploy:x:1001:1001:Deploy bot:/home/deploy:/bin/bash
```

Seven fields split by `:`: name, password placeholder, user ID, group ID, description, home directory, shell.

### Stream, don't slurp

A busy server can write a 40 GB access log in a day. You have two ways to read it:

```text
  slurp: read the whole file, then work       stream: one line at a time

  disk ──► [ 40 GB in memory ] ──► count       disk ──► [ 1 line ] ──► count
                                                         ▲    │
            needs 40 GB of RAM                           └────┘
            (the process gets killed)              needs a few KB of RAM
```

*Streaming* means processing data piece by piece as it arrives, keeping only what you need. Memory stays flat however big the file gets. It also works on data that never ends, like `tail -f` or a log stream from `kubectl logs -f`.

In Go, `bufio.Scanner` streams line by line:

```go
package main

import (
	"bufio"
	"fmt"
	"strings"
)

const accessLog = `10.0.0.7 - - [13/Oct/2026:09:15:02 +0000] "GET /health HTTP/1.1" 200 2
10.0.0.9 - - [13/Oct/2026:09:15:03 +0000] "POST /deploy HTTP/1.1" 502 157
broken line
10.0.0.4 - - [13/Oct/2026:09:15:04 +0000] "GET /login HTTP/1.1" 503 0
`

func main() {
	scanner := bufio.NewScanner(strings.NewReader(accessLog))
	lines, serverErrors, skipped := 0, 0, 0
	for scanner.Scan() {
		lines++
		fields := strings.Fields(scanner.Text())
		if len(fields) < 9 {
			skipped++
			continue
		}
		if strings.HasPrefix(fields[8], "5") {
			serverErrors++
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Println("read failed:", err)
		return
	}
	fmt.Printf("%d lines, %d server errors, %d skipped\n", lines, serverErrors, skipped)
}
```

It prints `4 lines, 2 server errors, 1 skipped`.

1. `scanner.Scan()` reads the next line and returns `false` at the end.
2. `strings.Fields` splits on runs of spaces. Field 8 (counting from 0) is the status code.
3. `broken line` has too few fields. Checking `len(fields)` first stops an index-out-of-range panic.
4. `scanner.Err()` tells you whether the loop stopped because of the end of input or a real read error.

Here the input is a string so the example is self-contained. With a real file you'd pass the open file to `bufio.NewScanner` instead, and the loop stays the same. One limit to know: by default a Scanner refuses lines longer than 64 KB and stops with an error, so `scanner.Err()` matters.

You build exactly this in *Reading input line by line* and the *Project: access log report*.

## Structured formats

Line formats are flat. Real configuration has nesting: a Deployment has containers, a container has ports. Structured formats handle that.

| Format | Looks like | Where you meet it |
|---|---|---|
| CSV | `web-1,10.0.0.5,eu-west-1` | inventories, exports, billing reports |
| JSON | `{"replicas": 3}` | APIs, `kubectl -o json`, AWS CLI output |
| YAML | `replicas: 3` | Kubernetes manifests, CI pipelines, Ansible |
| TOML / INI | `region = "eu-west-1"` | `~/.aws/config`, systemd units, containerd |

### CSV: not as simple as it looks

CSV means *comma-separated values*: one row per line, one comma between fields, often a header row first.

The trap is that fields may contain commas. A description like `web, primary` must be quoted: `"web, primary"`. So `strings.Split(line, ",")` is wrong for real CSV. Use a real CSV reader, as in *CSV files*.

### JSON: strict and everywhere

JSON has six kinds of values: objects `{}`, arrays `[]`, strings, numbers, `true`/`false`, and `null`.

```json
{"name": "checkout", "replicas": 3, "ports": [8080, 9090], "canary": false}
```

JSON is strict on purpose. No comments, no trailing commas, strings always in double quotes. That makes it a poor format for humans to edit, and a great one for programs to exchange.

### YAML: friendly and full of traps

YAML is what Kubernetes manifests are written in:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: checkout
spec:
  replicas: 3
```

It is easy to read, but its friendliness comes from guessing, and the guesses cause outages:

- **Indentation is structure.** Moving a line two spaces changes what it belongs to. Tabs are not allowed for indentation at all.
- **Implicit types.** YAML decides the type from how a value looks. `version: 1.10` is the *number* 1.1, not the string `"1.10"`.
- **The Norway problem.** Under the older YAML 1.1 rules, which many tools still follow, `NO`, `no`, `yes`, `on` and `off` are booleans. A list of country codes `[GB, FR, NO]` becomes `["GB", "FR", false]`.
- **Leading zeros.** Under YAML 1.1, `mode: 0644` is an octal (base 8) number, 420 in decimal. Kubernetes file-mode fields like `defaultMode: 0644` rely on this; elsewhere it surprises people.

The fix for all of them: **when you mean a string, quote it.** In a Kubernetes container, an environment variable written as `value: true` is rejected because env values must be strings; `value: "true"` works.

You decode real manifests in *YAML manifests*, in the *Kubernetes with client-go* module.

### TOML and INI: key = value

INI files group `key = value` lines under `[section]` headers. Your AWS CLI config is one:

```ini
[profile prod]
region = eu-west-1
output = json
```

TOML is a stricter, typed version of the same idea, used by containerd and many tools. Terraform's `terraform.tfvars` uses HCL, Terraform's own language, which looks similar: `instance_type = "t3.small"`.

## Parsing and serialization

Programs don't work on text; they work on values like structs, slices and numbers. So data crosses a border twice:

```text
              parse / unmarshal / decode
  bytes  ─────────────────────────────────►  Go value
 (JSON)  ◄─────────────────────────────────  (struct)
           serialize / marshal / encode
```

- *Parsing* turns bytes into a value. In Go's JSON package this is called *unmarshalling* or *decoding*.
- *Serialization* turns a value into bytes. Go calls it *marshalling* or *encoding*.

A parser checks *syntax*: is this valid JSON at all? It does not know whether `replicas: -5` makes sense for your service. That is the next step.

## Parse, then validate

*Validation* checks *meaning*: the data is well-formed, but is it allowed? Replicas must be at least 1, a port must be between 1 and 65535, a name must not be empty.

A *schema* is a written description of what valid data looks like: which fields exist, their types, their limits. Kubernetes has a schema for every kind of object, and that is how `kubectl apply` can reject `replicas: "three"` before anything runs.

This program does both steps on three config snippets:

```go
package main

import (
	"encoding/json"
	"errors"
	"fmt"
)

type Service struct {
	Name     string `json:"name"`
	Replicas int    `json:"replicas"`
	Port     int    `json:"port"`
}

func (s Service) Validate() error {
	var errs []error
	if s.Name == "" {
		errs = append(errs, errors.New("name is required"))
	}
	if s.Replicas < 1 {
		errs = append(errs, fmt.Errorf("replicas must be at least 1, got %d", s.Replicas))
	}
	if s.Port < 1 || s.Port > 65535 {
		errs = append(errs, fmt.Errorf("port must be 1-65535, got %d", s.Port))
	}
	return errors.Join(errs...)
}

func load(data string) {
	var s Service
	if err := json.Unmarshal([]byte(data), &s); err != nil {
		fmt.Println("parse error:", err)
		return
	}
	if err := s.Validate(); err != nil {
		fmt.Println("invalid:", err)
		return
	}
	out, _ := json.Marshal(s)
	fmt.Println("ok:", string(out))
}

func main() {
	load(`{"name": "checkout", "replicas": 3, "port": 8080}`)
	load(`{"name": "checkout", "replicas": 3, "port": 8080,}`)
	load(`{"name": "checkout", "replcias": 3, "port": 80800}`)
}
```

It prints:

```text
ok: {"name":"checkout","replicas":3,"port":8080}
parse error: invalid character '}' looking for beginning of object key string
invalid: replicas must be at least 1, got 0
port must be 1-65535, got 80800
```

Trace the three calls:

1. Valid syntax, valid values. `json.Marshal` turns the struct back into compact JSON.
2. The trailing comma before `}` is not allowed in JSON. The parser stops: a *syntax* error.
3. The syntax is fine, so parsing succeeds. But `replcias` is a typo. Go's JSON decoder quietly ignores keys it doesn't know, so `Replicas` keeps its zero value, 0. Only validation catches it, together with the impossible port.

That third case is the whole argument for validating. **Parsing tells you the data is well-formed. Validation tells you it's correct enough to act on.** Do it at the edge, before a bad config reaches a deploy. Joining every problem into one error, as in *Reporting many errors*, lets the user fix them all in one go.

For config files you can also make typos fail loudly: a `json.Decoder` with `DisallowUnknownFields()` turned on rejects `replcias` with an `unknown field` error. You still need validation for values that are spelled right but wrong, like the port.

You practise both directions in *Encoding JSON* and *Decoding JSON*. The *Errors and failure* guide explains why returning an error beats crashing here.

## Regular expressions

A *regular expression* (regex) is a small pattern language for matching text. `v\d+\.\d+\.\d+` means "a `v`, digits, a dot, digits, a dot, digits", so it matches release tags like `v1.27.3`.

Regex is great at one thing: **pulling fields out of a single line with a fixed shape.** Here it splits an nginx log line into named pieces:

```go
package main

import (
	"fmt"
	"regexp"
)

var logLine = regexp.MustCompile(`^(\S+) \S+ \S+ \[([^\]]+)\] "(\S+) (\S+) [^"]*" (\d{3}) (\d+)`)

func main() {
	line := `10.0.0.9 - - [13/Oct/2026:09:15:03 +0000] "POST /deploy HTTP/1.1" 502 157 "-" "curl/8.5.0"`
	m := logLine.FindStringSubmatch(line)
	if m == nil {
		fmt.Println("no match")
		return
	}
	fmt.Println("ip:    ", m[1])
	fmt.Println("time:  ", m[2])
	fmt.Println("method:", m[3])
	fmt.Println("path:  ", m[4])
	fmt.Println("status:", m[5])

	fmt.Println(logLine.MatchString("this is not an access log line"))
}
```

It prints:

```text
ip:     10.0.0.9
time:   13/Oct/2026:09:15:03 +0000
method: POST
path:   /deploy
status: 502
false
```

Each pair of parentheses is a *capture group*. `m[0]` is the whole match, `m[1]` the first group, and so on. `\S+` means "one or more non-space characters"; `[^\]]+` means "anything up to the closing `]`".

### Where regex fails

Regex can't count matching brackets. JSON, YAML and HTML nest: an object inside an array inside an object. A pattern that works on today's output breaks when a field moves to a new line or a value contains a `}`.

**Rule of thumb:** regex for one line with a known shape; a real parser for anything that nests. Never parse JSON, YAML or HTML with regex.

Also keep patterns small and test them. A regex that nobody can read is a regex nobody dares to fix. Go's `regexp` package does protect you from one classic danger: its matching time grows only in step with the input, so a crafted log line can't make it hang.

You write and fix patterns in *Regular expressions*.

## Timestamps and time zones

Look at these two lines from different servers:

```text
web-1: 10/11/2026 01:30  deploy started
web-2: 2026-10-11T01:31:00Z  deploy finished
```

Is the first one 10 November or 11 October? Which time zone? If clocks went back an hour that night, 01:30 happened twice. You can't line these events up, and lining up events is most of incident response.

The fix is two habits:

1. **Use RFC 3339** (a profile of ISO 8601): `2026-10-11T01:31:00Z`. The biggest unit comes first, so timestamps in the same zone sort correctly even as plain text.
2. **Log in UTC.** The `Z` at the end means UTC. Convert to local time only when showing it to a human.

Go parses a time with a *layout*: the reference moment `Mon Jan 2 15:04:05 MST 2006`, written the way your input writes it.

```go
package main

import (
	"fmt"
	"time"
)

func main() {
	const nginxLayout = "02/Jan/2006:15:04:05 -0700"

	t, err := time.Parse(nginxLayout, "13/Oct/2026:11:15:03 +0200")
	if err != nil {
		fmt.Println("bad timestamp:", err)
		return
	}
	fmt.Println(t.Format(time.RFC3339))
	fmt.Println(t.UTC().Format(time.RFC3339))

	later, _ := time.Parse(time.RFC3339, "2026-10-13T09:20:03Z")
	fmt.Println(later.Sub(t))
}
```

It prints:

```text
2026-10-13T11:15:03+02:00
2026-10-13T09:15:03Z
5m0s
```

1. The layout tells `time.Parse` where the day, month, year and offset sit in nginx's format.
2. The instant prints first with its `+02:00` offset, then in UTC: 11:15 in a `+02:00` zone is 09:15 UTC.
3. Subtracting two times gives a duration. Go compares instants, not wall-clock text, so the different zones don't matter.

You'll handle cert expiry and maintenance windows in *Time and durations*.

## Common mistakes

- **Splitting CSV on commas.** Quoted fields contain commas. Use a CSV reader.
- **Reading a huge log into memory.** Stream it line by line instead.
- **Indexing fields without checking the count.** One short line panics the whole job. Check `len(fields)` first.
- **Unquoted YAML strings.** `NO`, `on`, `1.10`, `0644` may not be what you think. Quote strings.
- **Parsing without validating.** Typos in keys parse fine and turn into zero values.
- **Regex on nested data.** Use a JSON or YAML parser.
- **Local time without an offset in logs.** Use RFC 3339 in UTC.
- **Ignoring `scanner.Err()`.** A read error looks just like the end of the file unless you check.

## Check yourself

1. `len("Malmö")` is 6 but the word has 5 letters. Why?
2. Your job reads a 30 GB log with one call that loads the whole file, and gets killed. What do you change, and why does it help?
3. A YAML list of country codes `[SE, DK, NO]` reaches your Go code as `["SE", "DK", false]`. What happened, and how do you fix the file?
4. `json.Unmarshal` returned no error, but the service started with 0 replicas. What is the likely cause, and which step should have caught it?
5. A teammate wants to pull the `image` field out of `kubectl get pod -o json` with a regex. What do you suggest instead, and why?
6. Two log lines read `03:10` and `03:05`, from servers in different time zones. Which happened first?

## Answers

1. `len` counts bytes, not characters. In UTF-8, `ö` is stored as two bytes, so 5 characters take 6 bytes.
2. Stream it with a `bufio.Scanner`, one line at a time. Memory then holds one line plus your counters, a few KB, no matter how big the file is.
3. The parser used YAML 1.1 rules, where `NO` is the boolean false (the Norway problem). Quote it: `"NO"`.
4. A misspelt key, like `replcias`. Go's JSON decoder ignores unknown keys, so the field kept its zero value. Validation (`replicas` must be at least 1) should have rejected it after parsing.
5. Decode the JSON with a real parser (or use `kubectl -o jsonpath`). JSON nests and can change layout; a regex that works today breaks on the next change.
6. You can't tell. Without a time zone or offset the times aren't comparable. That's why logs should use RFC 3339 in UTC.

## Key terms

| Term | Meaning |
|---|---|
| byte | a number from 0 to 255; all stored and sent data is made of them |
| format | an agreement about what a sequence of bytes means |
| encoding | the rule that maps characters to bytes, such as UTF-8 |
| UTF-8 | the standard encoding: ASCII in 1 byte, other characters in 2–4 |
| parsing | turning bytes or text into values a program can use |
| serialization | turning values into bytes or text to store or send |
| marshal / unmarshal | Go's names for serializing and parsing |
| streaming | processing data piece by piece instead of all at once |
| validation | checking parsed data is allowed, not just well-formed |
| schema | a description of the fields, types and limits of valid data |
| regular expression | a pattern language for matching text |
| capture group | a `( )` part of a regex whose matched text you can extract |
| RFC 3339 | the standard timestamp format, e.g. `2026-10-11T01:31:00Z` |
| UTC | the world's reference time zone; `Z` in a timestamp |

## Watch

- [Characters, Symbols and the Unicode Miracle - Computerphile](https://www.youtube.com/watch?v=MijmeoH9LT4) · Computerphile · 10 min: Tom Scott shows how UTF-8 turns characters into bytes, which is why `len("Malmö")` is 6.
- [The Problem with Time & Timezones - Computerphile](https://www.youtube.com/watch?v=-5wpm-gesOY) · Computerphile · 10 min: a funny rant on why time zones are a mess, and why you store timestamps in UTC.
- [Yaml Tutorial | Learn YAML in 18 mins](https://www.youtube.com/watch?v=1uFVr15xDGg) · TechWorld with Nana · 18 min: YAML syntax walked through with real Kubernetes manifests, the files you'll parse most in ops.
- [Learn Regular Expressions In 20 Minutes](https://www.youtube.com/watch?v=rhzKDrUiJVk) · Web Dev Simplified · 21 min: a beginner-friendly tour of regex patterns built up one piece at a time in a live tester.

## Go deeper

- [JSON and Go](https://go.dev/blog/json): the Go team's walkthrough of marshalling and unmarshalling, including how unknown fields are handled.
- [Strings, bytes, runes and characters in Go](https://go.dev/blog/strings): Rob Pike on how Go stores text as UTF-8 bytes.
- [The Absolute Minimum Every Software Developer Must Know About Unicode](https://www.joelonsoftware.com/2003/10/08/the-absolute-minimum-every-software-developer-absolutely-positively-must-know-about-unicode-and-character-sets-no-excuses/): a funny, classic explanation of why "plain text" doesn't exist.
- [The Norway Problem](https://hitchdev.com/strictyaml/why/implicit-typing-removed/): why the StrictYAML project removed implicit typing, with the Norway problem and other real examples of YAML guessing wrong.
- [RFC 3339: Date and Time on the Internet](https://www.rfc-editor.org/rfc/rfc3339): the timestamp standard; sections 4 and 5 explain why UTC and this ordering.
