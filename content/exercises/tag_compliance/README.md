# Tag compliance: audit and fix instance tags

An AWS account fills up with servers fast, and soon nobody knows who owns
`i-0b2f...` or whether it is safe to turn off. Tags answer that. In this
exercise you write a small audit tool that finds untagged EC2 instances and
prepares the fix, carefully.

## Learn

### What a tag is

A *tag* is a key/value label you attach to an AWS resource: an instance, a
bucket, a load balancer. AWS doesn't care what the tags say. They are for
you:

| Key | Example value | Answers the question |
|---|---|---|
| `Owner` | `payments-team` | Who do I page when this breaks? |
| `Env` | `prod` | Can I restart it on a Tuesday afternoon? |
| `CostCenter` | `CC-1042` | Whose budget pays for it? |

Most companies make some tags mandatory. *Tag policies* in AWS Organizations
can enforce allowed values, and the monthly *cost report* groups the bill by
tag, so an untagged server's cost lands in "unknown". Finance chases the
platform team about it, and the platform team writes an audit tool like this
one. Tag keys are **case-sensitive**: `owner` and `Owner` are two different
tags.

In the SDK an EC2 tag is a `types.Tag` from
`github.com/aws/aws-sdk-go-v2/service/ec2/types`. Like everything in the SDK,
its fields are pointers (see *SDK clients: small interfaces and fakes*):

```go
type Tag struct {
	Key   *string
	Value *string
}
```

An instance's tags are a `[]types.Tag` in `inst.Tags`. The easiest way to
look one up is to copy them into a map first. `aws.ToString` turns a `nil`
pointer into `""`, so a missing value is safe:

```go
have := make(map[string]string, len(inst.Tags))
for _, t := range inst.Tags {
	have[aws.ToString(t.Key)] = aws.ToString(t.Value)
}
have["Owner"] // "" if the tag is absent OR set to ""
```

### Recap: every instance, every page

As in *Paginators: walk every page*, `DescribeInstances` returns instances a
page at a time, grouped into *reservations*. The paginator loop reads all of
them:

```go
p := ec2.NewDescribeInstancesPaginator(api, &ec2.DescribeInstancesInput{})
for p.HasMorePages() {
	page, err := p.NextPage(ctx)
	if err != nil {
		return nil, fmt.Errorf("describe instances: %w", err)
	}
	for _, r := range page.Reservations {
		for _, inst := range r.Instances {
			// inst.InstanceId, inst.Tags ...
		}
	}
}
```

Pages arrive in whatever order AWS likes, so sort the report before you show
it, as in *Sorting*. `strings.Compare(a, b)` works like `cmp.Compare` for
strings:

```go
slices.SortFunc(findings, func(a, b Finding) int {
	return strings.Compare(a.InstanceID, b.InstanceID)
})
```

### Fixing tags: `CreateTags`, with a dry run first

`CreateTags` adds tags to one or more resources (it overwrites a tag with the
same key). You describe the change with an `*ec2.CreateTagsInput`:

```go
in := &ec2.CreateTagsInput{
	Resources: []string{"i-0b2"},
	Tags: []types.Tag{
		{Key: aws.String("Owner"), Value: aws.String("platform-team")},
	},
	DryRun: aws.Bool(true),
}
// later: client.CreateTags(ctx, in)
```

`DryRun` is built into most EC2 calls that change things. With
`aws.Bool(true)`, EC2 checks that you are allowed to make the change and
then changes **nothing**. It answers with a `DryRunOperation` error, which
here means "this would have worked" (as in *AWS API errors: EC2 dry-run
permission checks*).

Any tool that changes infrastructure should work this way: show the plan,
run it as a dry run, and only apply it when someone asks for that. Building
the requests in a separate function, as you will, also makes that function
easy to test without AWS.

## Your task

Your team must tag every instance with `CostCenter`, `Env` and `Owner`
(`RequiredTags` in `exercise.go`). Finish the audit and the fixer.

### 1. Find missing tags

Change `MissingTags` to return the keys from `RequiredTags` that are not in
`tags`, or that are there with an empty value. Keep `RequiredTags` order. A
compliant instance gives an empty result (`nil` is fine). The map `have` is
already built for you.

```go
MissingTags(/* Env=prod, owner=payments, CostCenter="" */)
// => ["CostCenter" "Owner"]   ("owner" is not "Owner")
```

### 2. Audit every instance

`Audit` already pages through all instances and wraps API errors with
`describe instances: `. Change it to:

- add a `Finding` only for instances where `MissingTags(inst.Tags)` returns
  at least one key, with that result as `Missing`;
- sort the findings by `InstanceID` before returning them. Add `"slices"`
  and `"strings"` to the imports.

```go
// page 1: i-0c3 (no Owner), i-0a1 (all tags)
// page 2: i-0b2 (no tags)
Audit(ctx, api)
// => [{i-0b2 [CostCenter Env Owner]} {i-0c3 [Owner]}]
```

### 3. Build the fixes

Change `BuildFixes` to make one `*ec2.CreateTagsInput` per finding:

- `Resources` is `[]string{finding.InstanceID}` (already done);
- `Tags` holds one `types.Tag` for every key in `finding.Missing` that has an
  entry in `defaults`, in `Missing` order, with the default as its value;
- `DryRun` is `aws.Bool(dryRun)`.

Skip a finding when none of its missing keys has a default. Nobody can guess
an owner, so a human has to deal with it.

```go
BuildFixes(
	[]Finding{{"i-01", []string{"CostCenter", "Owner"}}, {"i-02", []string{"Env"}}},
	map[string]string{"Owner": "platform-team", "CostCenter": "CC-1000"},
	true,
)
// => one request: Resources=[i-01]
//    Tags=[CostCenter=CC-1000 Owner=platform-team] DryRun=true
```

## Learn more

- [Tag your Amazon EC2 resources](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/Using_Tags.html)
- [Best practices for tagging AWS resources](https://docs.aws.amazon.com/whitepapers/latest/tagging-best-practices/tagging-best-practices.html)
- [ec2.CreateTagsInput](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/ec2#CreateTagsInput)
- [Package slices: SortFunc](https://pkg.go.dev/slices#SortFunc)
