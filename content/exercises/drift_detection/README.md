# Drift detection: a tiny terraform plan

Someone "just quickly" fixed something in the AWS console, and now production
no longer matches what your team wrote down. You build a tiny `terraform plan`:
it compares what you declared with what really exists in AWS and reports the
differences.

**New here:** desired vs actual state (drift), plan exit codes, reading a JSON
file into a slice of structs.

## Learn

### Desired state, actual state, drift

With *Infrastructure as Code* (IaC) you describe buckets and servers in files
kept in git, and a tool such as Terraform creates them. Those files are the
*desired state*. What AWS really has right now is the *actual state*.

When the two disagree, you have *drift*:

| What happened | Action the plan reports |
|---|---|
| Someone deleted a bucket by hand (desired, missing in AWS) | **create** |
| Someone edited a tag, or added `hotfix=true` during an incident | **update** |
| Someone launched a debug server and forgot it (in AWS, not desired) | **delete** |

It is the same comparison as in *Reconcile loops: desired vs actual*. A
controller fixes drift straight away; a *plan* only reports it, so a human can
review it first.

### Exit codes that mean something

Drift checks run in CI every night, and CI only looks at the exit code.
`terraform plan -detailed-exitcode` gives each code a meaning:

| Exit code | Meaning |
|---|---|
| 0 | No changes. AWS matches the desired state. |
| 1 | Error. The check itself failed. |
| 2 | Changes found: there is drift. |

Your `Plan` returns the report text **and** this code.

### Reading a JSON file into structs

Reading a file is two steps: get the bytes, then decode them. Both can fail.
Here is the same idea with a different type and prefix:

```go
data, err := os.ReadFile("hosts.json")
if err != nil {
	return nil, fmt.Errorf("read hosts: %w", err)
}
var hosts []Host
err = json.Unmarshal(data, &hosts) // & so Unmarshal can fill hosts
```

> **Why is it written like this?** Go requires the `&`: `Unmarshal` must change
> your variable, so it needs its address. Checking `err` right after each call
> is convention. The prefix text (`read hosts: `) is the author's choice, but
> here the tests fix it.

The `json:"kind"` tags on `Resource` say which JSON key fills which field, as in
*Encoding JSON: a deployment status report*. The desired file looks like this:

```json
[
  { "kind": "bucket", "name": "acme-logs", "tags": { "Env": "prod", "Owner": "platform" } }
]
```

### "Is this key in the map?"

A map lookup can return a second value, `ok`, that says whether the key exists:

```go
if _, ok := stock["apples"]; !ok {
	fmt.Println("out of apples")
}
```

> **Why is it written like this?** The two-value form is Go syntax. `_` throws
> away the value you don't need (Go requires every variable to be used). The
> name `ok` is convention.

### Building text line by line

A `strings.Builder` collects text; `fmt.Fprintf(&b, ...)` writes into it.
`strings.Join` glues a slice together with a separator:

```go
var b strings.Builder
fmt.Fprintf(&b, "* %s\n", "disk full")
strings.Join([]string{"a", "b", "c"}, " | ") // => "a | b | c"
b.String()                                    // => "* disk full\n"
```

### What is already done

| File | What it holds |
|---|---|
| `aws.go` | Complete. `Actual` gathers buckets (`ListBuckets` + `GetBucketTagging`; `NoSuchTagSet` means "no tags") and instances (the `DescribeInstances` paginator). Terminated instances are skipped. An instance's name comes from its `Name` tag, so `Name` is not in its `Tags`. |
| `exercise.go` | `Resource`, `ID()`, `Change`, and the helpers `byID` and `changedKeys`. You fill in the TODOs. |
| `testdata/` | `desired.json` (4 resources) and `broken.json` (invalid JSON). |

`r.ID()` joins kind and name: `Resource{Kind: "bucket", Name: "acme-logs"}.ID()`
is `"bucket/acme-logs"`. `byID(rs)` returns a `map[string]Resource` keyed by
`ID()`. `changedKeys(want, got)` returns the sorted tag keys that differ
(including keys only one side has), or nil:

```go
changedKeys(map[string]string{"Env": "prod"}, map[string]string{"Env": "dev", "hotfix": "true"})
// => []string{"Env", "hotfix"}
```

The tests use fake S3 and EC2 clients, so nothing talks to real AWS.

## Your task

The platform team wants a nightly drift check for the production account.
Finish `exercise.go`.

### 1. Load the desired state

**Goal:** `LoadDesired(path)` reads the JSON file at `path` and returns its
resources as a `[]Resource`.

- Read the file with `os.ReadFile`, decode with `json.Unmarshal`.
- If **either** step fails, return `nil` and the error wrapped with `%w`,
  prefixed with exactly `load desired state: `.
- Add the imports `encoding/json`, `fmt` and `os`.

```go
LoadDesired("testdata/desired.json")
// => 4 resources: bucket/acme-artifacts, bucket/acme-logs, instance/web-1, instance/web-2
LoadDesired("testdata/missing.json")
// => error "load desired state: open testdata/missing.json: no such file or directory"
LoadDesired("testdata/broken.json")
// => error starting with "load desired state: " (wrapping a *json.SyntaxError)
```

**Think it through:**
- Which two calls can fail, and where does each `if err != nil` go?
- In the Learn example, which part is the prefix you must change?

### 2. Find the differences

**Goal:** `Diff(desired, actual)` also reports **create** and **delete**
changes, not only updates.

The starter already builds `want` and `got` (maps by ID), has the update loop,
and sorts everything by ID at the end. Fill in the two `TODO(2)` spots:

- **create**: every ID in `want` that is not in `got` →
  `Change{Action: "create", ID: id}` (no `Keys`).
- **delete**: every ID in `got` that is not in `want` →
  `Change{Action: "delete", ID: id}` (no `Keys`).

When nothing differs, return an empty or nil slice.

```go
desired: bucket/acme-logs {Env: prod}, instance/web-2 {Env: prod, Owner: web}
actual:  instance/web-2 {Env: staging, Owner: alice}, bucket/old-tmp {}
// => [{create bucket/acme-logs []} {delete bucket/old-tmp []} {update instance/web-2 [Env Owner]}]
```

**Think it through:**
- Which map do you loop over for "create", and which one do you look the ID up in?
- The update loop already looks up an ID in `got`. What does it do with `ok`?

### 3. Render the plan

**Goal:** `Plan` returns the report text and the exit code 0, 1 or 2.

Loading errors are already handled. Fill in the rest:

1. **`Actual` fails:** return `"Error: " + <the error's text> + "\n"` and **1**,
   exactly like the `LoadDesired` error just above it.
2. **No changes:** return exactly `"No changes. AWS matches the desired state.\n"` and **0**.
3. **Changes:** one line per change, in the order `Diff` returned them, then a
   summary line. Return **2**.

| Action | Line |
|---|---|
| create | `+ <id>` |
| update | `~ <id> (tags: <keys joined with ", ">)` |
| delete | `- <id>` |

Summary: `Plan: <creates> to add, <updates> to change, <deletes> to destroy.`
Every line, including the last, ends with `\n`.

```text
+ bucket/acme-artifacts
- bucket/old-tmp
~ instance/web-2 (tags: Env, Owner)
Plan: 1 to add, 1 to change, 1 to destroy.
```

Only a `hotfix` tag added to web-1:

```text
~ instance/web-1 (tags: hotfix)
Plan: 0 to add, 1 to change, 0 to destroy.
```

Add the `strings` import.

**Think it through:**
- Which field of a `Change` decides the line's first character?
- You need three counts for the summary. When do you add 1 to each?
- Which text is fixed (`+ `, ` (tags: `, `)`) and which comes from the change?

## Learn more

- [Terraform: `terraform plan` and `-detailed-exitcode`](https://developer.hashicorp.com/terraform/cli/commands/plan)
- [AWS CloudFormation: Detect drift](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-stack-drift.html)
- [encoding/json.Unmarshal](https://pkg.go.dev/encoding/json#Unmarshal)
- [strings.Builder](https://pkg.go.dev/strings#Builder)
