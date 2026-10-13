# Drift detection: a tiny terraform plan

Someone "just quickly" fixed something in the AWS console, and now production no
longer matches what your team wrote down. In this exercise you build a tiny
`terraform plan`: it compares the infrastructure you declared with what really
exists in AWS and reports the differences.

## Learn

### Infrastructure as Code

With *Infrastructure as Code* (IaC) you describe your buckets, servers and
networks in files, keep those files in git, and let a tool create the real
thing. Terraform and AWS CloudFormation are the best-known tools. The files
are the *desired state*: what should exist. What AWS really has right now is
the *actual state*.

When the two disagree, you have *drift*. Typical causes:

| What happened | What the tool sees |
|---|---|
| Someone deleted a bucket by hand | in desired, missing in AWS: **create** |
| Someone edited a tag in the console, or added `hotfix=true` during an incident | tags differ: **update** |
| Someone launched a debug server and forgot it | in AWS, not desired: **delete** |

This is the same desired-vs-actual comparison as in *Reconcile loops: desired vs actual*. The
difference: a controller fixes drift straight away, while a *plan* only
reports it, so a human can review it first.

### `terraform plan -detailed-exitcode`

Drift checks usually run in CI every night. The CI job only looks at the exit
code, so Terraform has a flag that gives the code a meaning:

| Exit code | Meaning |
|---|---|
| 0 | Success, no changes. AWS matches the desired state. |
| 1 | Error. The check itself failed (bad file, AWS refused a call). |
| 2 | Success, but there are changes: drift was found. |

Your `Plan` returns the report text and this exit code.

### The pieces in this exercise

A resource is identified by its kind and name. `ID()` joins them:

```go
r := Resource{Kind: "bucket", Name: "acme-logs", Tags: map[string]string{"Env": "prod"}}
r.ID() // => "bucket/acme-logs"
```

`aws.go` is already complete. Its `Actual` function gathers the actual state
with the calls you know from *Paginators: walk every page* and *Tag compliance: audit and fix instance tags*: `ListBuckets`
plus `GetBucketTagging` for each bucket (the error code `NoSuchTagSet` just
means "no tags"), and the `DescribeInstances` paginator for instances. An
instance's name comes from its `Name` tag, so `Name` is not in its `Tags`. The
tests pass fakes for both clients, so nothing talks to real AWS.

`exercise.go` also gives you two helpers. `byID(rs)` returns a
`map[string]Resource` keyed by `r.ID()`. `changedKeys(want, got)` returns the
sorted tag keys whose values differ, including keys only one side has, or nil
when the tags match:

```go
changedKeys(map[string]string{"Env": "prod"}, map[string]string{"Env": "dev", "hotfix": "true"})
// => []string{"Env", "hotfix"}
```

The desired state is a JSON file, `testdata/desired.json`:

```json
[
  { "kind": "bucket", "name": "acme-logs", "tags": { "Env": "prod", "Owner": "platform" } }
]
```

The `json:"kind"` struct tags on `Resource` tell `encoding/json` which field
each key goes into, as in *Encoding JSON: a deployment status report*.

## Your task

The platform team wants a nightly drift check for the production account.
Finish `exercise.go`.

### 1. Load the desired state

Implement `LoadDesired(path)`. Read the file with `os.ReadFile` and decode it
with `json.Unmarshal` into a `[]Resource`. Wrap any error (from either call) with
`%w` and the prefix `load desired state: `:

```go
fmt.Errorf("load desired state: %w", err)
```

Add the imports `encoding/json`, `fmt` and `os`.

```go
LoadDesired("testdata/desired.json")
// => 4 resources: bucket/acme-artifacts, bucket/acme-logs, instance/web-1, instance/web-2
LoadDesired("testdata/missing.json")
// => error "load desired state: open testdata/missing.json: no such file or directory"
```

### 2. Find the differences

Finish `Diff(desired, actual)`. The "update" loop and the final sort are
already there: for every ID in both `want` and `got`, the update loop calls
`changedKeys(w.Tags, g.Tags)` and appends
`Change{Action: "update", ID: id, Keys: keys}` when any keys differ. Add:

- **create**: for every ID in `want` that is not in `got`, append
  `Change{Action: "create", ID: id}`.
- **delete**: for every ID in `got` that is not in `want`, append
  `Change{Action: "delete", ID: id}`.

When nothing differs, `Diff` returns an empty (or nil) slice.

```go
desired: bucket/acme-logs {Env: prod}, instance/web-2 {Env: prod, Owner: web}
actual:  instance/web-2 {Env: staging, Owner: alice}, bucket/old-tmp {}
// => [{create bucket/acme-logs []} {delete bucket/old-tmp []} {update instance/web-2 [Env Owner]}]
```

### 3. Render the plan

Finish `Plan`. It already loads the desired state and turns a load error into
`"Error: <err>\n"` with exit code 1.

- If `Actual` returns an error, return `"Error: " + err.Error() + "\n"` and **1**
  from the `if err != nil` block that is already there.
- If `Diff` finds no changes, return exactly this and **0**:

  ```text
  No changes. AWS matches the desired state.
  ```

- Otherwise write one line per change, in the order `Diff` returned them, then
  a summary line, and return **2**:

  | Action | Line |
  |---|---|
  | create | `+ <id>` |
  | update | `~ <id> (tags: <keys joined with ", ">)` |
  | delete | `- <id>` |

  The summary counts each action: `Plan: <creates> to add, <updates> to change, <deletes> to destroy.`

Every line, including the last one, ends with `\n`. For the drifted account in
the tests:

```text
+ bucket/acme-artifacts
- bucket/old-tmp
~ instance/web-2 (tags: Env, Owner)
Plan: 1 to add, 1 to change, 1 to destroy.
```

A `strings.Builder` with `fmt.Fprintf(&b, ...)` and `strings.Join(c.Keys, ", ")`
make this short. Add the `strings` import.

## Learn more

- [Terraform: `terraform plan` and `-detailed-exitcode`](https://developer.hashicorp.com/terraform/cli/commands/plan)
- [AWS CloudFormation: Detect drift](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-stack-drift.html)
- [aws-sdk-go-v2 ec2.NewDescribeInstancesPaginator](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/ec2#NewDescribeInstancesPaginator)
- [encoding/json.Unmarshal](https://pkg.go.dev/encoding/json#Unmarshal)
