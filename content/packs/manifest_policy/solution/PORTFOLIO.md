# kubepolicy

[![ci](https://github.com/YOUR-NAME/kubepolicy/actions/workflows/ci.yml/badge.svg)](https://github.com/YOUR-NAME/kubepolicy/actions/workflows/ci.yml)

A command-line tool that checks Kubernetes manifests against deployment
policy before they reach a cluster, and fails the CI build when they don't
comply.

## The problem

Most production incidents I read about started with a manifest that looked
fine in review: a container with no memory limit that starved its node of
memory, an image tagged `latest` that changed under a rollback, or a
production Deployment with 1 replica that went down during a routine node
drain. Reviewers miss these because they are absences, not mistakes you can
see. kubepolicy turns the team's rules into a check that runs on every pull
request, so a missing probe is a red build, not a 3 a.m. page.

## How it works

```text
files, dirs, stdin ─► manifest.Read ─► Split ─► Decode (typed k8s.io/api structs)
                                                     │
                     report.Table / JSON ◄── policy.Evaluate ◄── policy.Rules
```

`manifest.Read` collects YAML from files, directories and stdin, and `Split`
cuts each stream into documents at `---` lines. `Decode` reads each document
into the real Kubernetes Go type for its kind (Deployment, StatefulSet,
DaemonSet, Job or Pod), so a field with the wrong type is an error instead
of a guess; other kinds are skipped with a note. The rules are a table of small
values: a name, a severity and a check function. `Evaluate` runs every rule
over every object, so adding a rule is one function and one table entry, and
`--ignore` works for every rule for free. The report prints an aligned table
for people or JSON for scripts, and the exit code is 0 (clean), 1 (errors)
or 2 (bad input).

## Run it

```bash
go build -o bin/kubepolicy ./cmd/kubepolicy
bin/kubepolicy deploy/base                                   # a folder of manifests
kubectl kustomize deploy/overlays/production | bin/kubepolicy -   # what production gets
bin/kubepolicy -format json -ignore probes testdata/mixed.yaml
bin/kubepolicy -enable image-digest deploy/base              # opt-in: images pinned by digest
```

Exit codes: `0` no errors (warnings are fine), `1` at least one error,
`2` a bad command line or broken YAML.

## Trade-offs

- It checks files before they are applied, not requests inside the cluster, so it is simple to run anywhere but can't stop someone who runs kubectl apply by hand.
- Warnings, like missing labels, never fail the build, which keeps adoption easy but means some teams will ignore them.
- Typed decoding with the k8s.io/api structs catches wrong field types, but unknown fields are silently ignored, so a typo like livenesProbe shows up only as a missing probe.
- The rules require CPU limits, which some teams deliberately leave off to avoid throttling; they can pass -ignore resources, but that also drops the memory checks.

## What I'd do next

- Split the resources rule so CPU limits can be ignored without losing memory checks.
- Check init containers and the Pod Security Standards restricted profile, such as dropping all capabilities.
- Read rule settings like production namespaces and required labels from a config file.
- Run the same rules as a validating admission webhook, so the cluster enforces them too.
