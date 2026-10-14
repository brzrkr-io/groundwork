# Capstone: a policy checker for Kubernetes manifests

Every platform team has rules for what may run in its clusters: set
resource limits, add probes, don't run as root, don't deploy `:latest`.
You'll build `kubepolicy`, a command-line tool that checks Kubernetes YAML
against those rules in CI, and publish it on GitHub as a portfolio project.

**New here:** `bytes.Lines`, `kind: List`. Everything else is practice from
across the course, including multi-document YAML (*YAML manifests: reading
them into Go*), `securityContext` (*Security contexts: least privilege for
pods*) and property checks in a fuzz target (*Testing*).

## Learn

### What you are building

```console
$ kubepolicy testdata/mixed.yaml
kubepolicy: skipped testdata/mixed.yaml#1: unsupported kind "ConfigMap"
LOCATION               OBJECT                        RULE       SEVERITY  MESSAGE
testdata/mixed.yaml#3  production/StatefulSet/cache  resources  error     container "redis": missing limits.cpu, limits.memory
testdata/mixed.yaml#3  production/StatefulSet/cache  labels     warning   missing labels: team
...

3 objects checked, 5 errors, 1 warning
$ echo $?
1

$ kubectl kustomize deploy/overlays/production | kubepolicy -format json -
{
  "objects": 1,
  "skipped": 2,
  "violations": []
}
```

It reads files, directories (every `.yaml` and `.yml` file under them) and
stdin (`-`). It checks five kinds of workload: Deployment, StatefulSet,
DaemonSet, Job and Pod. Other kinds, such as Services, get a note on stderr
and are skipped.

### How the code is organised

```text
paths, stdin ─► manifest.Read ─► manifest.Split ─► manifest.Decode ─► []manifest.Object
                                                                            │
     stdout ◄── report.Table / report.JSON ◄── policy.Evaluate ◄── policy.Rules
                                                       │
                                        cli.Run picks the exit code
```

| File | What it holds | Tasks |
|---|---|---|
| `cmd/kubepolicy/main.go` | The thin `main`, which calls `cli.Run`. Done. | – |
| `internal/manifest/read.go` | `Document`, `Read` (files, directories, stdin, `kind: List`). Done. | – |
| `internal/manifest/split.go` | `Split`: one YAML stream into documents. | 1 |
| `internal/manifest/decode.go` | `Object`, `Decode`: a document into a workload. | 2 |
| `internal/manifest/split_fuzz_test.go` | `FuzzSplit` and `checkSplit`, which you write. | 10 |
| `internal/policy/policy.go` | `Rule`, the `Rules` table, `Violation`, `Evaluate`. | 4, 5 |
| `internal/policy/rules.go` | One check function per rule. | 3, 5 |
| `internal/report/report.go` | `Table` and `JSON`. | 6, 7 |
| `internal/cli/cli.go` | `Run`, its flags and its exit codes. | 8 |
| `internal/cli/golden_test.go` | `TestGolden`, which you finish. | 9 |
| `ci.yml` | The CI workflow. | 11 |
| `PORTFOLIO.md` | The README your GitHub repository will show. | 12 |
| `testdata/` | Manifests to try: `good.yaml`, `mixed.yaml`, `invalid.yaml`, `list.yaml`, and `violations/` with one file per rule. | – |
| `deploy/` | A small app's manifests, which your CI checks (a kustomize base and a production overlay). | – |
| `Dockerfile`, `Makefile` | The image, and shortcuts such as `make build`, `make test` and `make policy`. Done. | – |
| `*grader_test.go`, `workflow_test.go` | Groundwork's checks for tasks 9–12. Delete them when you publish. | – |

Work in task order: each package's tests are next to it, and the
`internal/cli` tests skip with "finish task N first" until the packages
below them work.

> **Why is it written like this?** Go requires `internal/` to mean "only
> this module may import these packages". `cmd/<name>/` is community
> convention for a program's `main`. The four packages, and which one owns
> what, are the author's choice: each can be tested alone, and the arrows
> in the diagram only point one way.

### Multi-document YAML

One YAML file (or one `kubectl kustomize` output) can hold many objects.
A line `---` separates them:

```yaml
# Source: shop/templates/web.yaml
---
kind: Deployment
...
--- # the Service in front of it
kind: Service
...
```

Here a *separator* is a line that starts with `---`, followed by nothing,
or by spaces and maybe a `#` comment. `----`, `---x` and an indented
`    ---` (inside a multi-line string) are text, not separators. The part
before the first separator is a document too, and here it holds only a
comment, so it doesn't count.

### Walking lines and keeping your place

`bytes.Lines(data)` gives you each line of a `[]byte` in turn, **with** its
`\n` (the last line may have none). Add up the lengths, and you always know
where the current line starts. A slice expression then cuts out a piece:

```go
data := []byte("v1.2\n- fix login\n\nv1.1\n- add search\n")
start, pos := 0, 0
for line := range bytes.Lines(data) {
	if bytes.HasPrefix(line, []byte("v")) && pos > 0 {
		fmt.Printf("%q\n", data[start:pos]) // the release notes so far
		start = pos
	}
	pos += len(line)
}
fmt.Printf("%q\n", data[start:])
// "v1.2\n- fix login\n\n"
// "v1.1\n- add search\n"
```

`data[start:pos]` is not a copy: it shares memory with `data`. The `bytes`
package has the same helpers as `strings`, for `[]byte`: `TrimSpace`,
`HasPrefix`, `CutPrefix`, `TrimRight`, and so on.

> **Why is it written like this?** `for line := range bytes.Lines(data)`
> works because `bytes.Lines` returns an *iterator* (a function `range`
> knows how to call; Go 1.24 added it). Keeping the `\n` in each line is
> the library's choice, so lengths add up to the whole input.

### Decoding by kind

As in *YAML manifests: decoding kinds and linting images*: read only the
`kind` into a `metav1.TypeMeta`, then decode again into the Go type for that
kind:

```go
var tm metav1.TypeMeta
yaml.Unmarshal(doc, &tm)
switch tm.Kind {
case "Secret":
	var s corev1.Secret
	err = yaml.Unmarshal(doc, &s)
	name = s.Name
case "ConfigMap":
	var c corev1.ConfigMap
	err = yaml.Unmarshal(doc, &c)
	name = c.Name
}
```

The rules don't care whether a pod came from a Deployment or a Job, so
`Decode` copies the parts they need into one shape, `manifest.Object`:
the name, namespace and labels, the replica count, and the pod spec. A
Deployment, StatefulSet, DaemonSet and Job keep their pod spec at
`.Spec.Template.Spec`; a Pod's is its own `.Spec`. Jobs live in package
`batchv1 "k8s.io/api/batch/v1"`.

`Replicas` is an `*int32`, so "not written" (`nil`) and `0` differ. The API
server fills in 1 when a Deployment or StatefulSet leaves it out, and the
starter's `defaultReplicas` does the same. DaemonSets, Jobs and Pods have no
replica count at all, so theirs stays `nil`.

### Rules as a table

As in *Lookup tables*, data beats a chain of `if`s. A rule is a small value,
and its check is a function stored in a struct field, as in *Functions as
values*:

```go
type check struct {
	name string
	fn   func(a Alert) []string
}

var checks = []check{
	{name: "runbook", fn: hasRunbook},
	{name: "owner", fn: func(a Alert) []string {
		if a.Owner == "" {
			return []string{"no owner"}
		}
		return nil
	}},
}

for _, c := range checks {
	for _, problem := range c.fn(alert) {
		fmt.Println(c.name, problem)
	}
}
```

Adding a check is adding an entry. Everything that loops over the table
(running checks, `--ignore`, listing names) works for the new one with no
other change.

### The fields the rules read

| Rule | Where to look | Unset means |
|---|---|---|
| resources | `c.Resources.Requests[corev1.ResourceCPU]`, `...ResourceMemory`, and the same in `c.Resources.Limits` | the key isn't in the map |
| probes | `c.LivenessProbe`, `c.ReadinessProbe` (`*corev1.Probe`) | `nil` |
| security-context | `obj.Pod.SecurityContext.RunAsNonRoot`, `c.SecurityContext.RunAsNonRoot`, `.AllowPrivilegeEscalation`, `.ReadOnlyRootFilesystem` (all `*bool`) | `nil`, or a `nil` `SecurityContext` |
| image-tag, image-digest | `c.Image` (a string) | – |
| labels | `obj.Labels` (a map) | missing key, or `""` |
| host-access | `obj.Pod.HostNetwork` (`bool`), `v.HostPath` for each `v` in `obj.Pod.Volumes` (`*corev1.HostPathVolumeSource`) | `false`, `nil` |
| replicas | `obj.Replicas`, `obj.Namespace` | `nil` |

`c` is each container in `obj.Pod.Containers`. A `*bool` has three states,
so check the pointer before reading through it:

```go
sc := c.SecurityContext
privileged := sc != nil && sc.Privileged != nil && *sc.Privileged
_, hasStorage := c.Resources.Limits[corev1.ResourceEphemeralStorage]
```

> **Why is it written like this?** Go evaluates `&&` left to right and stops
> at the first `false`, so `*sc.Privileged` only runs when both pointers are
> safe. The Kubernetes API uses pointers so "you didn't say" and "you said
> false" stay different; that's the API's choice.

### Exit codes

As in *Standard streams: grep-style exit codes*:

| Code | Meaning |
|---|---|
| `0` | no violation with severity `error` (warnings are fine) |
| `1` | at least one `error` |
| `2` | the tool couldn't do its job: a bad flag, a missing file, broken YAML, no documents |

A CI step fails on anything but `0`, and a script can still tell "your
manifests are wrong" from "I couldn't read them".

### Golden files and fuzzing, again

From *Test fixtures and golden files*: a golden file holds a program's
exact expected output, and an `-update` flag rewrites it after a change you
meant to make. From *Fuzzing*: a fuzz target feeds generated inputs to your
code and checks a *property*, something true for every input.

For a function that returns pieces of its input, put the properties in a
helper that returns an error. It's easy to test with broken results on
purpose, and the fuzz target stays short:

```go
// checkFields checks what strings.Fields returned for s.
func checkFields(s string, fields []string) error {
	for _, f := range fields {
		if f == "" || strings.ContainsAny(f, " \t\n") {
			return fmt.Errorf("field %q is empty or has a space", f)
		}
	}
	return nil
}

func FuzzFields(f *testing.F) {
	f.Add("web-1  web-2")
	f.Fuzz(func(t *testing.T, s string) {
		if err := checkFields(s, strings.Fields(s)); err != nil {
			t.Fatal(err)
		}
	})
}
```

## Your task

Your team keeps losing on-call nights to manifests that looked fine in
review. Build the check that runs on every pull request, then publish it.

### 1. Split a YAML stream

**Goal:** in `internal/manifest/split.go`, make `Split(data)` return the documents in `data`, without the separators and without empty documents.

- Each document is exactly the bytes between two separators (or between
  the start or end of `data` and a separator), line endings included.
- A separator is a line that starts with `---`, then has nothing more, or
  spaces or tabs and optionally a `#` comment (see *Multi-document YAML*).
  Windows line endings (`---\r\n`) count too.
- Leave out documents with nothing but blank lines and comment lines (lines
  whose first non-space character is `#`).
- Empty input gives no documents (`nil` or an empty slice).

```go
Split([]byte("# Source: web.yaml\n---\nkind: A\n--- # b\nkind: B"))
// => [][]byte{[]byte("kind: A\n"), []byte("kind: B")}

Split([]byte("kind: A\ndata:\n  script: |\n    ---\n"))
// => one document: the indented --- is text
```

`Read` (already written) calls `Split` for every file, and numbers the
documents from 1: `deploy/web.yaml#2` is the second document with content.
Until you finish this task, each whole file counts as one document.

**Think it through:**
- Keep two positions while you walk the lines: where the current document
  started, and where the current line starts. What happens to each at a
  separator?
- The last document has no separator after it. Where do you add it?
- Two small helpers, "is this line a separator?" and "does this document
  have content?", make `Split` short. You'll reuse them in task 10.

### 2. Decode the workloads

**Goal:** in `internal/manifest/decode.go`, make `Decode(doc)` handle StatefulSet, DaemonSet, Job and Pod, and return an `ErrUnsupportedKind` error for every other kind.

The `"Deployment"` case is written: copy its shape.

| Kind | Go type | `pod` comes from | `replicas` |
|---|---|---|---|
| `StatefulSet` | `appsv1.StatefulSet` | `.Spec.Template.Spec` | `defaultReplicas(s.Spec.Replicas)` |
| `DaemonSet` | `appsv1.DaemonSet` | `.Spec.Template.Spec` | stays `nil` |
| `Job` | `batchv1.Job` | `.Spec.Template.Spec` | stays `nil` |
| `Pod` | `corev1.Pod` | `.Spec` | stays `nil` |

Any other kind, including a document without one, returns `Object{}` and an
error that wraps `ErrUnsupportedKind` with `%w`, in exactly this form:

```go
Decode(Document{Source: "deploy/app.yaml", Index: 3, Data: []byte("kind: Service\n")})
// => Object{}, error `deploy/app.yaml#3: unsupported kind "Service"`
```

Build it with `fmt.Errorf`: `doc.Location()`, `": "`, the wrapped
`ErrUnsupportedKind`, a space, and the kind in quotes (`%q`). Broken YAML
already returns an error starting with the location; keep that. Add the
import `batchv1 "k8s.io/api/batch/v1"`.

**Think it through:**
- What does each case set? Look at the three variables the `"Deployment"`
  case assigns, and at the table.
- Which `switch` keyword catches every kind you didn't list?
- Why must the error *wrap* `ErrUnsupportedKind` rather than just say the
  same words? Look at how `load` in `cli.go` tells skips from real errors.

### 3. Write the first rule

**Goal:** in `internal/policy/rules.go`, make `checkResources(obj)` return one message per container that is missing a CPU or memory request or limit.

- Check every container in `obj.Pod.Containers`, in order.
- Check four entries, in this order: `requests.cpu`, `requests.memory`,
  `limits.cpu`, `limits.memory`. An entry is missing when its key isn't in
  the map.
- One message per container that misses any, listing what it misses,
  joined with `", "`. A container that has all four adds nothing.
- Add the imports `fmt`, `strings` and `corev1 "k8s.io/api/core/v1"`.

```text
container "web": missing requests.cpu, requests.memory, limits.cpu, limits.memory
container "log-shipper": missing limits.memory
```

The `Rules` table in `policy.go` already lists this rule:
`{Name: "resources", Severity: Error, Check: checkResources}`.

**Think it through:**
- `corev1.ResourceCPU` is a map key. How do you ask a map whether it has a
  key, as in *Maps*?
- Collect the missing names for one container in a slice. When do you turn
  that slice into a message, and when do you start a fresh one?

### 4. Run the rule table

**Goal:** in `internal/policy/policy.go`, make `Evaluate(objects, cfg)` run every selected rule over every object and return the violations.

- Order: objects in the order given; for each object, the rules in `Rules`
  order; for each rule, its messages in the order it returned them.
- Skip a rule whose name is in `cfg.Ignore`.
- Skip a rule with `OptIn: true` unless its name is in `cfg.Enable`. When a
  name is in both lists, ignore wins.
- Each message becomes a `Violation` with the object's `Location`, `Kind`,
  `Namespace` and `Name`, the rule's `Name` (in the `Rule` field) and
  `Severity`, and the message.
- No violations returns `nil` (or an empty slice).
- Add the import `slices`.

```go
Evaluate([]manifest.Object{web}, Config{})
// => []Violation{{Location: "deploy/web.yaml#1", Kind: "Deployment",
//      Namespace: "production", Name: "web", Rule: "resources",
//      Severity: "error", Message: `container "web": missing limits.cpu`}}
```

The test swaps in a small rule table of its own, so it doesn't need task 5.

**Think it through:**
- Three loops, one inside the other. Which list does each loop walk?
- Write the two "skip this rule" conditions as named booleans before the
  `continue`. Does it read like the rules above?

### 5. Add the remaining rules

**Goal:** write a check function for each rule below, and add each one to `Rules` in this order, so the table holds all eight rules.

| Name | Severity | Opt-in | Messages |
|---|---|---|---|
| `resources` | `Error` | no | (task 3) |
| `probes` | `Error` | no | `container "web": missing livenessProbe, readinessProbe` |
| `security-context` | `Error` | no | `container "web": needs runAsNonRoot: true, allowPrivilegeEscalation: false, readOnlyRootFilesystem: true` |
| `image-tag` | `Error` | no | `container "web": image "nginx" has no tag` / `container "web": image "nginx:latest" uses the latest tag` |
| `image-digest` | `Error` | **yes** | `container "web": image "nginx:1.27" is not pinned by digest` |
| `labels` | `Warning` | no | `missing labels: app.kubernetes.io/name, team` |
| `host-access` | `Error` | no | `hostNetwork is true` / `volume "logs" uses hostPath "/var/log"` |
| `replicas` | `Error` | no | `replicas is 1, want at least 2 in namespace "production"` |

The details, rule by rule (every container means every container in
`obj.Pod.Containers`, in order):

- **probes:** per container, list what's missing, `livenessProbe` before
  `readinessProbe`, like task 3. A `Job` returns `nil`: its containers run
  to completion, so probes make no sense.
- **security-context:** per container, list what it needs, in the order
  shown. `runAsNonRoot` comes from the container's `securityContext` when
  it sets it, otherwise from the pod's (`obj.Pod.SecurityContext`); it
  must be `true`. `allowPrivilegeEscalation` (container only) must be set
  to `false`. `readOnlyRootFilesystem` (container only) must be set to
  `true`.
- **image-tag:** per container. An image with `@` in it is pinned by
  digest and always passes. Otherwise the tag is the text after the last
  `:`, but only if that `:` comes after the last `/`
  (`registry.example.com:5000/ops/shipper` has no tag), as in *YAML
  manifests: decoding kinds and linting images*. No tag gives the first
  message, tag `latest` the second.
- **image-digest:** per container: an image without `@sha256:` in it.
- **labels:** each label in `RequiredLabels` that is missing or empty in
  `obj.Labels`, in that order, in one message.
- **host-access:** first `hostNetwork is true` if `obj.Pod.HostNetwork`
  is set, then one message per volume with a `HostPath`, in order, with its
  name and path.
- **replicas:** only when `obj.Replicas` isn't `nil` and `obj.Namespace` is
  in `ProductionNamespaces`: a count below 2 gives the message, with the
  count and the namespace.

A rule with no problems returns `nil`. The fixtures in `testdata/violations/`
break one rule each; try them with
`go run ./cmd/kubepolicy testdata/violations/probes.yaml` once task 6 prints
a table.

**Think it through:**
- Most rules are task 3 again with different fields. Which ones look at
  the whole object instead of each container?
- For `runAsNonRoot`, start from the pod's answer and let the container
  overwrite it. What's the pod's answer when it says nothing?
- Is a `*bool` that's `nil` "set to `false`"? Look at *The fields the rules
  read*.

### 6. Print the report table

**Goal:** in `internal/report/report.go`, make `Table(w, rep)` print the violations as an aligned table, then a blank line, before the summary line it already prints.

- Use `tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)`, as in *Project: ops
  inventory CLI*.
- Header: `LOCATION`, `OBJECT`, `RULE`, `SEVERITY`, `MESSAGE`.
- One row per violation, in order. `OBJECT` is `v.Object()`, which is
  already written: `production/Deployment/web`, or `DaemonSet/log-agent`
  without a namespace.
- Flush the tabwriter (return its error), then print an empty line.
- No violations: no header and no blank line, only the summary (that part
  already works).
- Add the import `text/tabwriter`.

```text
LOCATION             OBJECT                     RULE         SEVERITY  MESSAGE
deploy/web.yaml#1    production/Deployment/web  resources    error     container "web": missing limits.cpu
deploy/agent.yaml#2  DaemonSet/log-agent        host-access  error     hostNetwork is true

3 objects checked, 2 errors, 0 warnings
```

**Think it through:**
- The empty line goes to `w`, not the tabwriter. Why does that matter for
  the order of the output?
- Which check decides whether to print a table at all?

### 7. Print JSON

**Goal:** make `JSON(w, rep)` print the report as one JSON object, indented with two spaces, with `"violations": []` when there are none.

- Use a `json.Encoder` writing to `w`, with `SetIndent("", "  ")`. The
  struct tags on `Report` and `Violation` already give the keys.
- `nil` violations must print `[]`, not `null`.
- Replace the `import "errors"` with `"encoding/json"`.

```json
{
  "objects": 4,
  "skipped": 0,
  "violations": []
}
```

**Think it through:**
- `rep` is a copy. Can you change its `Violations` before encoding without
  changing the caller's report?

### 8. Wire up flags and exit codes

**Goal:** in `internal/cli/cli.go`, add the repeatable `-ignore` and `-enable` flags, reject a bad `-format`, and make `Run` return `1` when there's an error-severity violation.

`-format` and the "no manifests" check are written. Add:

| Flag | Fills | Rejects |
|---|---|---|
| `-ignore RULE` (repeatable) | `opts.config.Ignore` | a name that isn't in `policy.Names()` |
| `-enable RULE` (repeatable) | `opts.config.Enable` | the same |

- Use `fs.Func` for each, as in *Flags: repeatable and validated flags*.
  When the function returns an error, `flag` prints it with the usage text
  and `Run` returns `2`. The error must contain `unknown rule "probe"`
  (for the name `probe`); listing the real rule names after it helps the
  user.
- After `fs.Parse`, reject a `-format` other than `table` or `json` with
  `usageError` and the message `-format must be table or json, got "xml"`.
- At the end of `Run`, return `exitViolations` when any violation has
  severity `policy.Error`; otherwise `exitClean`.
- Go's `flag` accepts `-ignore` and `--ignore` alike. Flags come before
  the paths.

```console
$ kubepolicy -ignore probes testdata/violations/probes.yaml; echo $?
1 object checked, no violations
0
$ kubepolicy -ignore probe testdata/good.yaml; echo $?
invalid value "probe" for flag -ignore: unknown rule "probe"; rules: resources, probes, ...
2
```

**Think it through:**
- Both flags need the same check, writing to different slices. Could one
  function that takes a `*[]string` and returns a `func(string) error`
  serve both (a closure, as in *Functions as values: state that sticks*)?
- Warnings alone exit `0`. Which field of each violation decides?

### 9. Lock the output with golden files

**Goal:** in `internal/cli/golden_test.go`, finish `TestGolden`: with `-update` it writes each case's output to its golden file, and it always fails when the output differs from the golden file.

`goldenCases`, the loop, the `Run` call and the `update` flag are written.
Replace the `t.Fatalf` line:

1. If `*update` is set: create `testdata/` with `os.MkdirAll` (it doesn't
   exist yet), and write `stdout.String()` to `golden` with
   `os.WriteFile`.
2. Read `golden` with `os.ReadFile`; stop with `t.Fatalf` if that fails.
3. If it differs from `stdout.String()`, fail with `t.Errorf` and a message
   that shows both.

Then create the golden files from the exercise folder, and read them:

```console
go test ./internal/cli -run TestGolden -update
```

`TestGrader` in `internal/cli/grader_test.go` checks your golden files
hold today's output, then re-runs your `TestGolden` in a scratch copy where
a golden file is wrong (it must fail) and with `-update` (it must write the
files).

**Think it through:**
- After `-update` writes the file, why read it back and compare anyway?
- From now on, a change to the output fails this test. When that change is
  on purpose, what do you run, and what do you check before you commit?

### 10. Fuzz the splitter

**Goal:** in `internal/manifest/split_fuzz_test.go`, write `checkSplit` and finish `FuzzSplit` so the fuzzer checks every result of `Split`.

`checkSplit(data, docs)` returns `nil` when `docs` is a correct split of
`data`, and otherwise an error saying which property broke:

1. Every document has content: a line that isn't blank and isn't a comment.
2. No document has a separator line inside it.
3. Every document is a piece of `data`, in order: document 2 comes after
   document 1, and so on (`bytes.Index` on the rest of `data` after the
   previous document).
4. Nothing is lost: the documents together have as many *content lines*
   (not blank, not a comment, not a separator) as `data`.

In `FuzzSplit`:

- Add at least three seeds with `f.Add`, each a `[]byte`, such as a stream
  with a comment header and two documents.
- Inside `f.Fuzz`, replace the `t.Fatalf` line: call `Split(data)`, pass
  the result to `checkSplit`, and `t.Fatal` with its error.

```go
checkSplit([]byte("kind: A\n---\nkind: B\n"), [][]byte{[]byte("kind: A\n")})
// => error: the documents have 1 content lines, the input has 2
```

`go test` replays your seeds and the four inputs in
`testdata/fuzz/FuzzSplit/` (Windows line endings, separators with
comments, a separator at the very end, dashes that are text). If one fails,
the bug is in `Split` or in `checkSplit`: the failure names the input.
`TestGrader` feeds `checkSplit` some broken results on purpose and checks
it catches each one. Then let the fuzzer loose: `make fuzz`.

**Think it through:**
- Your helpers from task 1 already answer "is this a separator?" and "does
  this have content?". Which properties do they cover?
- For property 3, keep a slice `rest` that starts as `data`. What do you
  cut off it after each document you find?

### 11. Check every push in CI

**Goal:** in `ci.yml`, write the workflow `ci` with two jobs: `test` vets and tests the code; `policy` runs kubepolicy on the `deploy/` manifests.

- `name: ci`. Triggers: `push` limited to the branch `main`, and
  `pull_request`.
- Job `test`, on `ubuntu-latest`, steps in order: `actions/checkout` and
  `actions/setup-go` (setup-go with `go-version-file: go.mod`),
  `go vet ./...`, `go test -race ./...`. As in
  *Practice: containerize and ship a service*.
- Job `policy`, on `ubuntu-latest`, with `needs: test`. Steps: the same
  pinned checkout and setup-go, then
  - a step that runs kubepolicy on `deploy/base`, and
  - a step that pipes `kubectl kustomize deploy/overlays/production` into
    kubepolicy with `-`.

  Run kubepolicy with `go run ./cmd/kubepolicy`, or build it first
  (`go build -o bin/kubepolicy ./cmd/kubepolicy`) and run
  `bin/kubepolicy`. `make policy` does both checks too. `kubectl` is already
  installed on GitHub's Ubuntu runners.

```yaml
      - run: kubectl kustomize deploy/overlays/production | bin/kubepolicy -
```

Pin every action as in *CI workflows*: the full commit SHA, then
`# vX.Y.Z`. The test checks every `uses` line. The pins:
`actions/checkout@3d3c42e5aac5ba805825da76410c181273ba90b1 # v7.0.1` and
`actions/setup-go@b7ad1dad31e06c5925ef5d2fc7ad053ef454303e # v7.0.0`.

**Think it through:**
- Why check the overlay's *rendered* output, not only the files in
  `deploy/`? Which rule can only fire after kustomize adds the namespace?
- The step's exit code is the last command's in the pipe. Which program is
  last, and what does its exit code mean?

### 12. Write the portfolio README

**Goal:** fill in every section of `PORTFOLIO.md`, the README your GitHub repository will show.

The `<!-- comments -->` say what each section needs; they don't show on
GitHub, and the test ignores them. Replace every `TODO` and keep the
headings.

| Section | The test wants |
|---|---|
| `## The problem` | at least 40 words |
| `## How it works` | at least 60 words |
| `## Run it` | a fenced code block (between two lines of three backticks) with a `kubepolicy` command in it |
| `## Trade-offs` | at least 2 `- ` bullet points, each at least 6 words |
| `## What I'd do next` | at least 3 `- ` bullet points, each at least 4 words |

Also replace the `TODO` sentence under the title. The test can't tell
whether your writing is good; a reader can. Write for a hiring manager who
skims: lead with what the tool does and why it matters, then show you know
what you traded away.

**Think it through:**
- Which incident does each rule prevent? That's your problem section.
- Where did you choose between two reasonable options (warnings that don't
  fail, typed structs, skipping unknown kinds)? Each one is a trade-off.

When all twelve tests pass, try *Run it for real*, then *Publish it*.

## Run it for real

Optional, and worth it: you'll run kubepolicy on real manifests, as an
image, and against a live cluster. Run everything from the exercise folder.

**1. Build it and try the fixtures.**

```bash
make build
bin/kubepolicy testdata/violations; echo "exit $?"
bin/kubepolicy -format json testdata/mixed.yaml
bin/kubepolicy -ignore probes -enable image-digest testdata/good.yaml
```

**2. Check what kustomize renders**, the way your CI will:

```bash
kubectl kustomize deploy/overlays/production | bin/kubepolicy -
```

**3. Check the go links service.** If you did *Capstone: go links, built
and shipped*, its folder is next to this one:

```bash
kubectl kustomize ../*_golinks/deploy/overlays/dev | bin/kubepolicy -
```

Expect findings: that project left out a CPU limit on purpose, and its
labels aren't the ones your rules want. A good trade-off for your
portfolio: is your rule wrong, or that manifest?

**4. Build the image** and run it with the manifests on stdin (`-i`):

```bash
docker build -t kubepolicy:dev .
kubectl kustomize deploy/overlays/production | docker run --rm -i kubepolicy:dev -
```

**5. Start a local cluster** with kind (*Kubernetes in Docker*).
`nix shell` opens a shell with `kind` in it; run the remaining steps in
that shell.

```bash
nix shell nixpkgs#kind
kind create cluster --name gw-kubepolicy
```

If it fails with `could not find a log line that matches "Reached target
.*Multi-User System.*"`, Docker logs to journald and `docker logs` can't
read it back. Add `virtualisation.docker.logDriver = "json-file";` to your
NixOS configuration, run `sudo nixos-rebuild switch`, and try again.

**6. Deploy the app your policy approved**, and talk to it:

```bash
kubectl apply -k deploy/overlays/production
kubectl -n production rollout status deployment/hello
kubectl -n production port-forward service/hello 8080:80 &
sleep 2   # give the port-forward a moment to start
curl localhost:8080
# => hello from a policy-checked pod
kill %1
```

It runs as non-root, on a read-only root filesystem, with probes and
limits: the rules weren't just paperwork.

**7. Audit what's running.** `kubectl get -o yaml` prints a `kind: List`,
and `Read` unpacks it:

```bash
kubectl get deployments,daemonsets,statefulsets -A -o yaml | bin/kubepolicy -
```

Your `hello` passes. The cluster's own parts (`kindnet`, `kube-proxy`)
break `host-access`, because a network plugin *must* touch the node. Real
policy tools let you exempt namespaces like `kube-system`; that's an idea
for *What I'd do next*.

**8. Run kubepolicy inside the cluster.** Load your image into the kind
node, then run it as a one-off pod with your terminal's stdin attached:

```bash
kind load docker-image kubepolicy:dev --name gw-kubepolicy
kubectl kustomize deploy/overlays/production | kubectl run kubepolicy -i --rm --restart=Never --image=kubepolicy:dev --image-pull-policy=Never -- -
```

**9. Clean up.**

```bash
kind delete cluster --name gw-kubepolicy
docker rmi kubepolicy:dev
exit   # leaves the nix shell
```

## Publish it

1. On GitHub, create an empty public repository named `kubepolicy` (no
   README, no licence: you have those files).
2. Copy the project out of the course and make it a normal repository:

   ```bash
   cp -r . ~/code/kubepolicy && cd ~/code/kubepolicy
   mv PORTFOLIO.md README.md
   mkdir -p .github/workflows && mv ci.yml .github/workflows/ci.yml
   mv Dockerfile.dockerignore .dockerignore
   rm grader_test.go workflow_test.go internal/*/grader_test.go
   printf 'bin/\n' > .gitignore
   ```

   Your actions are pinned to commit SHAs, so nothing updates them by
   itself. Let Dependabot open a pull request when a new release comes out;
   it changes the SHA and the `# vX.Y.Z` comment together:

   ```bash
   cat > .github/dependabot.yml <<'EOF'
   version: 2
   updates:
     - package-ecosystem: github-actions
       directory: /
       schedule:
         interval: weekly
   EOF
   ```

3. Change the module path to your repository (replace `YOUR-NAME` with
   your GitHub user name), then check everything still builds:

   ```bash
   go mod edit -module github.com/YOUR-NAME/kubepolicy
   grep -rl 'groundwork.local/manifest_policy' --include='*.go' . | xargs sed -i 's#groundwork.local/manifest_policy#github.com/YOUR-NAME/kubepolicy#g'
   go vet ./... && go test ./...
   ```

4. In `README.md`, replace both `YOUR-NAME`s in the badge line.
5. Commit and push:

   ```bash
   git init -b main
   git add .
   git commit -m "kubepolicy: a policy checker for Kubernetes manifests"
   git remote add origin git@github.com:YOUR-NAME/kubepolicy.git
   git push -u origin main
   ```

6. Open the repository's **Actions** tab and watch the `ci` run: `test`,
   then `policy`, both green. The badge in your README turns green too. If
   a job fails, its log shows the step and the output, as on your machine.
7. On your GitHub profile, choose **Customize your pins** and pin
   `kubepolicy`, so it's the first thing a visitor sees.

## Learn more

- [Kubernetes: Configure a security context for a pod or container](https://kubernetes.io/docs/tasks/configure-pod-container/security-context/)
- [Kubernetes: Pod Security Standards](https://kubernetes.io/docs/concepts/security/pod-security-standards/)
- [Kubernetes: Configure liveness, readiness and startup probes](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/)
- [kind: Quick start](https://kind.sigs.k8s.io/docs/user/quick-start/)
- [pkg.go.dev: bytes.Lines](https://pkg.go.dev/bytes#Lines)
