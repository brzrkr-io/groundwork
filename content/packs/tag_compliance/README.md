# Tag compliance: audit and fix instance tags

An AWS account fills up with servers fast, and soon nobody knows who owns
`i-0b2f...` or whether it is safe to turn off. Tags answer that. Here you
write an audit tool that finds untagged EC2 instances and prepares the fix,
carefully.

**New here:** EC2 tags (`types.Tag`), `strings.Compare`, `CreateTags` requests with `DryRun`.

## Learn

### What a tag is

A *tag* is a key/value label on an AWS resource. AWS doesn't care what tags
say; they are for you:

| Key | Example value | Answers the question |
|---|---|---|
| `Owner` | `payments-team` | Who do I page when this breaks? |
| `Env` | `prod` | Can I restart it on a Tuesday afternoon? |
| `CostCenter` | `CC-1042` | Whose budget pays for it? |

Companies make some tags mandatory, and the monthly *cost report* groups the
bill by tag. An untagged server's cost lands in "unknown". Tag keys are
**case-sensitive**: `owner` and `Owner` are two different tags.

### Tags in the SDK

An EC2 tag is a `types.Tag` from
`github.com/aws/aws-sdk-go-v2/service/ec2/types`. Like everything in the SDK,
its fields are pointers (see *SDK clients: small interfaces and fakes*):

```go
type Tag struct {
	Key   *string
	Value *string
}
```

An instance's tags are a `[]types.Tag` in `inst.Tags`. To look one up, copy
them into a map first. `aws.ToString` turns a `nil` pointer into `""`:

```go
have := make(map[string]string, len(inst.Tags))
for _, t := range inst.Tags {
	have[aws.ToString(t.Key)] = aws.ToString(t.Value)
}
have["Team"] // "" if the tag is absent OR set to ""
```

> **Why is it written like this?** Reading a map key that isn't there is
> allowed in Go and gives the *zero value*, `""` for strings. That is a Go
> rule, and it lets one check cover "absent" and "empty".

### Recap: every instance, every page

As in *Paginators: walk every page*, `DescribeInstances` returns instances a
page at a time, grouped into *reservations*. The starter's `Audit` already
has the loop: pages, then reservations, then instances.

### Sorting by a string field

Pages arrive in whatever order AWS likes, so sort a report before showing it,
as in *Sorting*. `strings.Compare(a, b)` works like `cmp.Compare` for
strings: negative, zero or positive.

```go
type Volume struct {
	Name   string
	SizeGB int
}

slices.SortFunc(vols, func(a, b Volume) int {
	return strings.Compare(a.Name, b.Name)
})
```

> **Why is it written like this?** The compare function's shape,
> `func(a, b T) int`, is required by `slices.SortFunc`. The names `a` and `b`
> are convention.

### Fixing tags: CreateTags, with a dry run first

`CreateTags` adds tags to resources (overwriting a tag with the same key).
You describe the change with an `*ec2.CreateTagsInput`:

```go
in := &ec2.CreateTagsInput{
	Resources: []string{"vol-0f9"},
	Tags: []types.Tag{
		{Key: aws.String("Team"), Value: aws.String("storage")},
	},
	DryRun: aws.Bool(false),
}
// later: client.CreateTags(ctx, in)
```

With `DryRun` set to `aws.Bool(true)`, EC2 checks you're allowed and changes
**nothing** (as in *AWS API errors: EC2 dry-run permission checks*). Tools
that change infrastructure should show the plan, dry-run it, and only apply
on request. Building requests in their own function also makes them easy to
test without AWS.

## Your task

Your team must tag every instance with `CostCenter`, `Env` and `Owner`
(`RequiredTags` in `exercise.go`). Finish the audit and the fixer.

### 1. Find missing tags

**Goal:** `MissingTags(tags)` returns the required keys that are absent or empty, in `RequiredTags` order.

- A key counts as missing if it isn't in `tags`, or is there with an empty
  (or `nil`) value.
- Keys are case-sensitive: `owner` does not count as `Owner`.
- A compliant instance gives an empty result (`nil` is fine).

The map `have` is already built for you.

```go
MissingTags(/* Env=prod, owner=payments, CostCenter="" */)
// => ["CostCenter" "Owner"]   ("owner" is not "Owner")
```

**Think it through:**
- Which list do you loop over: the instance's tags, or `RequiredTags`?
- What does `have` give you for a key that isn't there?

### 2. Audit every instance

**Goal:** `Audit` returns one `Finding` per non-compliant instance, sorted by `InstanceID`.

It already pages through all instances and wraps API errors as
`describe instances: <err>`. Change it to:

- add a `Finding` only when `MissingTags` returns at least one key for the
  instance, and put that result in `Missing`;
- sort the findings by `InstanceID` before returning. Add `"slices"` and
  `"strings"` to the imports.

On an API error it returns `nil` findings (already done).

```go
// page 1: i-0c3 (no Owner), i-0a1 (all tags)
// page 2: i-0b2 (no tags)
Audit(ctx, api)
// => [{i-0b2 [CostCenter Env Owner]} {i-0c3 [Owner]}]
```

**Think it through:**
- Which instances should be skipped, and how do you know from `MissingTags`'s result?
- Does sorting happen inside the paginator loop, or after it?

### 3. Build the fixes

**Goal:** `BuildFixes` returns one `*ec2.CreateTagsInput` per finding it can fix, with defaults filled in.

For each finding:

- `Resources` holds the finding's instance ID (already done).
- `Tags` holds one `types.Tag` for each key in the finding's `Missing` that
  has an entry in `defaults`, in `Missing` order. The tag's value is that
  default.
- `DryRun` is set from the `dryRun` parameter.

Skip a finding when none of its missing keys has a default: nobody can guess
an owner, so a human must deal with it. No fixable findings gives `nil`.

```go
BuildFixes(
	[]Finding{{"i-01", []string{"CostCenter", "Owner"}}, {"i-02", []string{"Env"}}},
	map[string]string{"Owner": "platform-team", "CostCenter": "CC-1000"},
	true,
)
// => one request: Resources=[i-01]
//    Tags=[CostCenter=CC-1000 Owner=platform-team] DryRun=true
```

**Think it through:**
- How does a two-value map lookup (`value, ok`) tell "no default" apart from a default?
- After the inner loop, how do you know no tag was added?

## Learn more

- [Tag your Amazon EC2 resources](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/Using_Tags.html)
- [Best practices for tagging AWS resources](https://docs.aws.amazon.com/whitepapers/latest/tagging-best-practices/tagging-best-practices.html)
- [ec2.CreateTagsInput](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/ec2#CreateTagsInput)
- [Package slices: SortFunc](https://pkg.go.dev/slices#SortFunc)
