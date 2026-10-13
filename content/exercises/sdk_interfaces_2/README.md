# SDK clients: spy on StopInstances

A job that *stops servers* is one you really want tested before it runs. Here
you build a fake that records what it was asked to do, then use it to catch
the bugs in a nightly shutdown job.

**New here:** spies (fakes that record calls), `StopInstances`, spreading a slice into `append`.

## Learn

### Recap

From *SDK clients: small interfaces and fakes*: accept a small interface with
just the SDK methods you call, check the real client with
`var _ API = (*ec2.Client)(nil)`, and pass a fake in tests. SDK fields are
pointers: read them with `aws.ToString`, because `*field` panics on `nil`.

### Fakes that record: spies

A fake that only *answers* is enough for read calls. For a call that
*changes* something, the test also wants to know what your code asked for. A
fake that writes down its calls is a *spy*:

```go
type RecordingMailer struct {
	Sent []string
}

func (m *RecordingMailer) Send(to string) error {
	m.Sent = append(m.Sent, to)
	return nil
}
```

> **Why is it written like this?** The pointer receiver `(m *RecordingMailer)`
> is required: with a plain value receiver, the method would change a copy
> and the test would never see `Sent` grow. The name `m` is convention (a
> short abbreviation of the type).

### Appending a whole slice

`append(list, x)` adds one element. To add every element of another slice,
put `...` after it:

```go
all := []string{"a"}
more := []string{"b", "c"}
all = append(all, more...) // ["a" "b" "c"]
```

Go requires the `...`; without it, Go tries to add the slice as one element
and refuses to compile.

### StopInstances

`StopInstances` shuts down EC2 instances. You pass all IDs in one call, in
`ec2.StopInstancesInput.InstanceIds`. That field is a plain `[]string`, not
pointers. AWS rejects the call if the list is empty.

### Skipping items with continue

`continue` jumps straight to the next round of a loop. It keeps the "skip
this one" check at the top, so the rest of the loop body stays flat:

```go
for _, f := range files {
	if f.Size == 0 {
		continue
	}
	total += f.Size
}
```

## Your task

To save money, a nightly job stops every *running* instance tagged `Env=dev`.
It hasn't run in production yet, and it has bugs.

### 1. Finish the spy

**Goal:** `RecordingStopper.StopInstances` records every call, then returns
`r.Err` if it is set.

In this order:

- Add 1 to `r.Calls`.
- Append every ID in `params.InstanceIds` to `r.Stopped`.
- If `r.Err` is set, return `nil` and `r.Err`. Otherwise return
  `&ec2.StopInstancesOutput{}` and `nil`.

A failing call is still counted and its IDs are still recorded.

```go
r := &RecordingStopper{}
r.StopInstances(ctx, &ec2.StopInstancesInput{InstanceIds: []string{"i-web-1", "i-web-2"}})
r.StopInstances(ctx, &ec2.StopInstancesInput{InstanceIds: []string{"i-db-1"}})
// r.Calls   => 2
// r.Stopped => ["i-web-1" "i-web-2" "i-db-1"]
```

**Think it through:**
- Why must the recording happen *before* the error check?
- `params.InstanceIds` is a slice: how do you add all of it at once?

### 2. Stop `TagValue` from panicking

**Goal:** `TagValue(tags, key)` returns the value of the tag named `key`, or
`""`, without panicking on `nil` fields.

Right now it reads `*t.Key` and `*t.Value`, which panic when either is `nil`.
A tag with a `nil` key never matches. A matching tag with a `nil` value gives
`""`. Add the import `"github.com/aws/aws-sdk-go-v2/aws"`.

```go
tags := []ec2types.Tag{{Key: aws.String("Env"), Value: aws.String("prod")}, {Key: aws.String("Owner")}}
TagValue(tags, "Env")   // => "prod"
TagValue(tags, "Owner") // => "" (Value is nil)
TagValue(tags, "Team")  // => "" (no such tag)
```

**Think it through:**
- Which two spots use `*` on a pointer field?
- Which `aws` helper reads a `*string` and gives `""` for `nil`?

### 3. Fix `StopDevInstances`

**Goal:** `StopDevInstances(ctx, api, instances)` stops only *running*
instances tagged `Env=dev`, in one call, and never calls the API with an
empty list.

It picks the instances, sorts their IDs, stops them in one call and returns
the sorted IDs. Fix two bugs:

- It picks every `Env=dev` instance, even stopped ones. Only pick instances
  whose `State` is not `nil` and whose `State.Name` is
  `ec2types.InstanceStateNameRunning`.
- With nothing to stop, it still calls `StopInstances`. Return `nil, nil`
  instead, without calling the API.

The error wrapping (`stop instances: <err>`, with `%w`) is already right; on
error it returns `nil` and the wrapped error. The tests use your
`RecordingStopper`, so finish task 1 first.

```go
// i-dev-2 running Env=dev, i-prod-1 running Env=prod, i-dev-4 stopped Env=dev
StopDevInstances(ctx, spy, instances)
// => ["i-dev-2"], nil;  spy.Calls => 1, spy.Stopped => ["i-dev-2"]

// i-prod-1 running Env=prod
StopDevInstances(ctx, spy, instances)
// => nil, nil;  spy.Calls => 0
```

**Think it through:**
- Where in the loop is the best place to skip instances that aren't running?
- Between sorting and calling the API, what can you check about `ids`?

## Learn more

- [pkg.go.dev: service/ec2 Client.StopInstances](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/ec2#Client.StopInstances)
- [Amazon EC2: Stop and start instances](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/Stop_Start.html)
- [pkg.go.dev: aws.ToString](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/aws#ToString)
- [Go by Example: Slices](https://gobyexample.com/slices)
