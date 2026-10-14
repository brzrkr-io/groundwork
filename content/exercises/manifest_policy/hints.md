## Task 1: Split a YAML stream

Walk the lines with `bytes.Lines`, adding up their lengths so you know where
each line starts. At a separator, the bytes from the start of the current
document up to this line are one document; the next one starts after it.

### Closer

Keep `start, pos := 0, 0`. In the loop: if the line is a separator, add
`data[start:pos]` and set `start = pos + len(line)`; then `pos += len(line)`.
After the loop, add `data[start:]`. "Add" means "append, unless it has no
content". A separator check: `bytes.CutPrefix(bytes.TrimRight(line, "\r\n"),
[]byte("---"))`, then look at what's left.

### Answer

```go
func Split(data []byte) [][]byte {
	var docs [][]byte
	start, pos := 0, 0
	for line := range bytes.Lines(data) {
		if isSeparator(line) {
			docs = appendDoc(docs, data[start:pos])
			start = pos + len(line)
		}
		pos += len(line)
	}
	return appendDoc(docs, data[start:])
}

func appendDoc(docs [][]byte, doc []byte) [][]byte {
	if !hasContent(doc) {
		return docs
	}
	return append(docs, doc)
}

func isSeparator(line []byte) bool {
	rest, ok := bytes.CutPrefix(bytes.TrimRight(line, "\r\n"), []byte("---"))
	if !ok {
		return false
	}
	if len(rest) == 0 {
		return true
	}
	if rest[0] != ' ' && rest[0] != '\t' {
		return false
	}
	rest = bytes.TrimLeft(rest, " \t")
	return len(rest) == 0 || rest[0] == '#'
}

func hasContent(doc []byte) bool {
	for line := range bytes.Lines(doc) {
		text := bytes.TrimSpace(line)
		if len(text) > 0 && text[0] != '#' {
			return true
		}
	}
	return false
}
```

## Task 2: Decode the workloads

Each new case is the `"Deployment"` case with a different type and a
different place for the pod spec. The `default` case is the only one that
returns early.

### Closer

`case "Pod":` declares `var p corev1.Pod`, unmarshals into `&p`, and sets
`meta, pod = p.ObjectMeta, p.Spec`. The default case:
`return Object{}, fmt.Errorf("%s: %w %q", doc.Location(), ErrUnsupportedKind, tm.Kind)`.

### Answer

```go
case "StatefulSet":
	var s appsv1.StatefulSet
	err = yaml.Unmarshal(doc.Data, &s)
	meta, replicas, pod = s.ObjectMeta, defaultReplicas(s.Spec.Replicas), s.Spec.Template.Spec
case "DaemonSet":
	var d appsv1.DaemonSet
	err = yaml.Unmarshal(doc.Data, &d)
	meta, pod = d.ObjectMeta, d.Spec.Template.Spec
case "Job":
	var j batchv1.Job
	err = yaml.Unmarshal(doc.Data, &j)
	meta, pod = j.ObjectMeta, j.Spec.Template.Spec
case "Pod":
	var p corev1.Pod
	err = yaml.Unmarshal(doc.Data, &p)
	meta, pod = p.ObjectMeta, p.Spec
default:
	return Object{}, fmt.Errorf("%s: %w %q", doc.Location(), ErrUnsupportedKind, tm.Kind)
```

## Task 3: Write the first rule

Loop over the containers. For each one, start an empty list of missing
names, check the four map keys in order, and add a message only if the list
isn't empty.

### Closer

`_, ok := c.Resources.Requests[corev1.ResourceCPU]` tells you whether the
key is there; `!ok` means missing. Build the message with
`fmt.Sprintf("container %q: missing %s", c.Name, strings.Join(missing, ", "))`.

### Answer

```go
var problems []string
for _, c := range obj.Pod.Containers {
	var missing []string
	if _, ok := c.Resources.Requests[corev1.ResourceCPU]; !ok {
		missing = append(missing, "requests.cpu")
	}
	if _, ok := c.Resources.Requests[corev1.ResourceMemory]; !ok {
		missing = append(missing, "requests.memory")
	}
	if _, ok := c.Resources.Limits[corev1.ResourceCPU]; !ok {
		missing = append(missing, "limits.cpu")
	}
	if _, ok := c.Resources.Limits[corev1.ResourceMemory]; !ok {
		missing = append(missing, "limits.memory")
	}
	if len(missing) > 0 {
		problems = append(problems, fmt.Sprintf("container %q: missing %s", c.Name, strings.Join(missing, ", ")))
	}
}
return problems
```

## Task 4: Run the rule table

Outer loop: objects. Middle loop: `Rules`. Inner loop: the messages one
rule returns for one object. Decide whether to skip a rule before calling
its `Check`.

### Closer

`ignored := slices.Contains(cfg.Ignore, rule.Name)` and
`notEnabled := rule.OptIn && !slices.Contains(cfg.Enable, rule.Name)`; skip
with `continue` when either is true. Then
`for _, msg := range rule.Check(obj)` appends a `Violation{...}`.

### Answer

```go
var found []Violation
for _, obj := range objects {
	for _, rule := range Rules {
		ignored := slices.Contains(cfg.Ignore, rule.Name)
		notEnabled := rule.OptIn && !slices.Contains(cfg.Enable, rule.Name)
		if ignored || notEnabled {
			continue
		}
		for _, msg := range rule.Check(obj) {
			found = append(found, Violation{
				Location:  obj.Location,
				Kind:      obj.Kind,
				Namespace: obj.Namespace,
				Name:      obj.Name,
				Rule:      rule.Name,
				Severity:  rule.Severity,
				Message:   msg,
			})
		}
	}
}
return found
```

## Task 5: Add the remaining rules

`probes`, `security-context`, `image-tag` and `image-digest` loop over the
containers like task 3. `labels`, `host-access` and `replicas` look at the
object once. Write one function per rule, then one table entry each.

### Closer

For `runAsNonRoot`: `podNonRoot := false`, set it from
`obj.Pod.SecurityContext.RunAsNonRoot` when both aren't `nil`, then let a
non-`nil` container `RunAsNonRoot` overwrite it. A `*bool` that must be
`false` fails when `p == nil || *p`. For the tag:
`colon := strings.LastIndex(image, ":")`; there's no tag when
`colon <= strings.LastIndex(image, "/")`. The table entry for the opt-in
rule: `{Name: "image-digest", Severity: Error, OptIn: true, Check: checkImageDigest}`.

### Answer

```go
func checkProbes(obj manifest.Object) []string {
	if obj.Kind == "Job" {
		return nil
	}
	var problems []string
	for _, c := range obj.Pod.Containers {
		var missing []string
		if c.LivenessProbe == nil {
			missing = append(missing, "livenessProbe")
		}
		if c.ReadinessProbe == nil {
			missing = append(missing, "readinessProbe")
		}
		if len(missing) > 0 {
			problems = append(problems, fmt.Sprintf("container %q: missing %s", c.Name, strings.Join(missing, ", ")))
		}
	}
	return problems
}

func checkSecurityContext(obj manifest.Object) []string {
	podNonRoot := false
	if psc := obj.Pod.SecurityContext; psc != nil && psc.RunAsNonRoot != nil {
		podNonRoot = *psc.RunAsNonRoot
	}
	var problems []string
	for _, c := range obj.Pod.Containers {
		sc := c.SecurityContext
		if sc == nil {
			sc = &corev1.SecurityContext{}
		}
		nonRoot := podNonRoot
		if sc.RunAsNonRoot != nil {
			nonRoot = *sc.RunAsNonRoot
		}
		var needs []string
		if !nonRoot {
			needs = append(needs, "runAsNonRoot: true")
		}
		if sc.AllowPrivilegeEscalation == nil || *sc.AllowPrivilegeEscalation {
			needs = append(needs, "allowPrivilegeEscalation: false")
		}
		if sc.ReadOnlyRootFilesystem == nil || !*sc.ReadOnlyRootFilesystem {
			needs = append(needs, "readOnlyRootFilesystem: true")
		}
		if len(needs) > 0 {
			problems = append(problems, fmt.Sprintf("container %q: needs %s", c.Name, strings.Join(needs, ", ")))
		}
	}
	return problems
}

func checkImageTag(obj manifest.Object) []string {
	var problems []string
	for _, c := range obj.Pod.Containers {
		if strings.Contains(c.Image, "@") {
			continue
		}
		switch imageTag(c.Image) {
		case "":
			problems = append(problems, fmt.Sprintf("container %q: image %q has no tag", c.Name, c.Image))
		case "latest":
			problems = append(problems, fmt.Sprintf("container %q: image %q uses the latest tag", c.Name, c.Image))
		}
	}
	return problems
}

func imageTag(image string) string {
	colon := strings.LastIndex(image, ":")
	if colon <= strings.LastIndex(image, "/") {
		return ""
	}
	return image[colon+1:]
}

func checkImageDigest(obj manifest.Object) []string {
	var problems []string
	for _, c := range obj.Pod.Containers {
		if !strings.Contains(c.Image, "@sha256:") {
			problems = append(problems, fmt.Sprintf("container %q: image %q is not pinned by digest", c.Name, c.Image))
		}
	}
	return problems
}

func checkLabels(obj manifest.Object) []string {
	var missing []string
	for _, label := range RequiredLabels {
		if obj.Labels[label] == "" {
			missing = append(missing, label)
		}
	}
	if len(missing) == 0 {
		return nil
	}
	return []string{"missing labels: " + strings.Join(missing, ", ")}
}

func checkHostAccess(obj manifest.Object) []string {
	var problems []string
	if obj.Pod.HostNetwork {
		problems = append(problems, "hostNetwork is true")
	}
	for _, v := range obj.Pod.Volumes {
		if v.HostPath != nil {
			problems = append(problems, fmt.Sprintf("volume %q uses hostPath %q", v.Name, v.HostPath.Path))
		}
	}
	return problems
}

func checkReplicas(obj manifest.Object) []string {
	if obj.Replicas == nil || !slices.Contains(ProductionNamespaces, obj.Namespace) {
		return nil
	}
	if *obj.Replicas >= 2 {
		return nil
	}
	return []string{fmt.Sprintf("replicas is %d, want at least 2 in namespace %q", *obj.Replicas, obj.Namespace)}
}
```

And in `policy.go`:

```go
var Rules = []Rule{
	{Name: "resources", Severity: Error, Check: checkResources},
	{Name: "probes", Severity: Error, Check: checkProbes},
	{Name: "security-context", Severity: Error, Check: checkSecurityContext},
	{Name: "image-tag", Severity: Error, Check: checkImageTag},
	{Name: "image-digest", Severity: Error, OptIn: true, Check: checkImageDigest},
	{Name: "labels", Severity: Warning, Check: checkLabels},
	{Name: "host-access", Severity: Error, Check: checkHostAccess},
	{Name: "replicas", Severity: Error, Check: checkReplicas},
}
```

## Task 6: Print the report table

Wrap the table in `if len(rep.Violations) > 0`. Inside: a tabwriter, the
header, one row per violation, `Flush`, then a blank line straight to `w`.

### Closer

Cells are separated by `\t`: `fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n",
v.Location, v.Object(), v.Rule, v.Severity, v.Message)`. Return early if
`tw.Flush()` returns an error.

### Answer

```go
if len(rep.Violations) > 0 {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "LOCATION\tOBJECT\tRULE\tSEVERITY\tMESSAGE")
	for _, v := range rep.Violations {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n", v.Location, v.Object(), v.Rule, v.Severity, v.Message)
	}
	if err := tw.Flush(); err != nil {
		return err
	}
	fmt.Fprintln(w)
}
```

## Task 7: Print JSON

A `nil` slice encodes as `null`, an empty one as `[]`. Swap it before
encoding; `rep` is your own copy.

### Closer

`if rep.Violations == nil { rep.Violations = []policy.Violation{} }`, then
`enc := json.NewEncoder(w)`, `enc.SetIndent("", "  ")`, `return enc.Encode(rep)`.

### Answer

```go
func JSON(w io.Writer, rep Report) error {
	if rep.Violations == nil {
		rep.Violations = []policy.Violation{}
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(rep)
}
```

## Task 8: Wire up flags and exit codes

`fs.Func(name, usage, fn)` calls `fn` once per time the flag appears. One
helper can build the `fn` for both flags if it takes a pointer to the slice
it should fill.

### Closer

`func ruleList(names *[]string) func(string) error` returns a function that
checks `slices.Contains(policy.Names(), name)`, returns an error if not, and
otherwise does `*names = append(*names, name)`. Register it with
`fs.Func("ignore", "skip a rule (repeatable)", ruleList(&opts.config.Ignore))`.
For the exit code, loop over `rep.Violations` and look at `v.Severity`.

### Answer

```go
fs.Func("ignore", "skip `rule` (repeatable)", ruleList(&opts.config.Ignore))
fs.Func("enable", "also run the opt-in `rule` (repeatable)", ruleList(&opts.config.Enable))

// after the "no manifests" check:
if opts.format != "table" && opts.format != "json" {
	return options{}, usageError(fs, "-format must be table or json, got %q", opts.format)
}

// at the end of Run:
for _, v := range rep.Violations {
	if v.Severity == policy.Error {
		return exitViolations
	}
}
return exitClean

// ruleList returns a flag function that adds a rule name to *names, and
// rejects names that aren't in the rule table.
func ruleList(names *[]string) func(string) error {
	return func(name string) error {
		if !slices.Contains(policy.Names(), name) {
			return fmt.Errorf("unknown rule %q; rules: %s", name, strings.Join(policy.Names(), ", "))
		}
		*names = append(*names, name)
		return nil
	}
}
```

## Task 9: Lock the output with golden files

Two steps replace the `t.Fatalf`: write the file when `*update` is set, then
always read it and compare. Afterwards run the `-update` command once and
read what it wrote.

### Closer

`os.MkdirAll("testdata", 0o755)` and
`os.WriteFile(golden, []byte(stdout.String()), 0o644)` inside
`if *update { ... }`. Then `want, err := os.ReadFile(golden)` and
`if stdout.String() != string(want)`.

### Answer

```go
if *update {
	if err := os.MkdirAll("testdata", 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(golden, []byte(stdout.String()), 0o644); err != nil {
		t.Fatal(err)
	}
}
want, err := os.ReadFile(golden)
if err != nil {
	t.Fatalf("reading the golden file: %v (run with -update to create it)", err)
}
if stdout.String() != string(want) {
	t.Errorf("kubepolicy %s printed\n%s\nwant (from %s)\n%s",
		strings.Join(tc.args, " "), stdout.String(), golden, want)
}
```

Then: `go test ./internal/cli -run TestGolden -update`.

## Task 10: Fuzz the splitter

`checkSplit` is one loop over `docs` checking properties 1–3, then one
comparison for property 4. `hasContent` and `isSeparator` from task 1 do
most of the work; a `contentLines` helper counts lines for property 4.

### Closer

Keep `rest := data`. For each doc: `at := bytes.Index(rest, doc)`; `at < 0`
is an error, otherwise `rest = rest[at+len(doc):]`. Add up
`contentLines(doc)` and compare with `contentLines(data)` after the loop.
In `FuzzSplit`: `f.Add([]byte("kind: A\n---\nkind: B\n"))` and, inside,
`if err := checkSplit(data, Split(data)); err != nil { t.Fatal(err) }`.

### Answer

```go
func FuzzSplit(f *testing.F) {
	f.Add([]byte("kind: A\n---\nkind: B\n"))
	f.Add([]byte("# Source: web.yaml\n---\nkind: A\n--- # b\n\nkind: B"))
	f.Add([]byte("---\n# only a comment\n---\n"))
	f.Fuzz(func(t *testing.T, data []byte) {
		if err := checkSplit(data, Split(data)); err != nil {
			t.Fatal(err)
		}
	})
}

func checkSplit(data []byte, docs [][]byte) error {
	rest := data
	lines := 0
	for i, doc := range docs {
		if !hasContent(doc) {
			return fmt.Errorf("document %d has no content: %q", i+1, doc)
		}
		for line := range bytes.Lines(doc) {
			if isSeparator(line) {
				return fmt.Errorf("document %d has a separator line inside: %q", i+1, doc)
			}
		}
		at := bytes.Index(rest, doc)
		if at < 0 {
			return fmt.Errorf("document %d is not a piece of the input after document %d: %q", i+1, i, doc)
		}
		rest = rest[at+len(doc):]
		lines += contentLines(doc)
	}
	if want := contentLines(data); lines != want {
		return fmt.Errorf("the documents have %d content lines, the input has %d", lines, want)
	}
	return nil
}

func contentLines(data []byte) int {
	n := 0
	for line := range bytes.Lines(data) {
		text := bytes.TrimSpace(line)
		if len(text) > 0 && text[0] != '#' && !isSeparator(line) {
			n++
		}
	}
	return n
}
```

The file's imports become `"bytes"`, `"fmt"` and `"testing"`.

## Task 11: Check every push in CI

The `test` job is the one from *Practice: containerize and ship a
service*. The `policy` job starts the same way, then builds kubepolicy and
runs it twice.

### Closer

`needs: test` goes at the job's level, next to `runs-on`. The two checks are
`run: bin/kubepolicy deploy/base` and
`run: kubectl kustomize deploy/overlays/production | bin/kubepolicy -`, after a
`run: go build -o bin/kubepolicy ./cmd/kubepolicy` step.

### Answer

```yaml
name: ci

on:
  push:
    branches: [main]
  pull_request:

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v7
      - uses: actions/setup-go@v7
        with:
          go-version-file: go.mod
      - run: go vet ./...
      - run: go test -race ./...

  policy:
    needs: test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v7
      - uses: actions/setup-go@v7
        with:
          go-version-file: go.mod
      - run: go build -o bin/kubepolicy ./cmd/kubepolicy
      - run: bin/kubepolicy deploy/base
      - run: kubectl kustomize deploy/overlays/production | bin/kubepolicy -
```

## Task 12: Write the portfolio README

Write each section as if the reader has never heard of the project and has
two minutes. Your own answers to tasks 1–11 are the material: what each rule
prevents, how the packages fit together, what you'd change.

### Closer

Problem: name a concrete failure and its cost. How it works: follow one
manifest from `Read` to the exit code, and say why the rules are a table.
Run it: paste the commands from *Run it for real*. Trade-offs: "I chose X,
which costs Y", one per bullet.

### Answer

The reference solution's `PORTFOLIO.md` (shown when you finish) is one
complete example. Yours should be in your own words, about your choices:
for example, a trade-off like "Warnings never fail the build, which keeps
adoption easy but means some teams will ignore them."
