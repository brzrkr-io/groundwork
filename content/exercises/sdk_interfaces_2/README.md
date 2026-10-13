# SDK clients: spy on StopInstances

A job that *stops servers* is one you really want tested before it runs. Here
you build a fake that records what it was asked to do, then use it to catch
the bugs in a nightly shutdown job.

## Learn

**Recap.** Accept a small interface with just the SDK methods you call, check
the real client with `var _ API = (*ec2.Client)(nil)`, and pass a fake in
tests. SDK fields are pointers: read them with `aws.ToString`, never with `*`.

### Fakes that record: spies

The fakes in *SDK clients: small interfaces and fakes* answered questions.
For a call that *changes* something, such as stopping a server, the test also
wants to know what your code asked for. A fake that writes down its calls is
a *spy*:

```go
spy := &RecordingStopper{}
StopDevInstances(ctx, spy, instances)
// spy.Calls   => 1
// spy.Stopped => ["i-dev-1" "i-dev-2"]
```

Because the spy's method has a pointer receiver, it can change `r.Calls` and
`r.Stopped`, and the test sees the changes afterwards.

### StopInstances

`StopInstances` shuts down EC2 instances. You pass the IDs in one call:

```go
_, err := api.StopInstances(ctx, &ec2.StopInstancesInput{
	InstanceIds: []string{"i-dev-1", "i-dev-2"},
})
```

`InstanceIds` is a plain `[]string`, not pointers. AWS rejects the call if the
list is empty.

## Your task

To save money, a nightly job stops every *running* instance tagged `Env=dev`.
It hasn't run in production yet, and it has bugs.

### 1. Finish the spy

Make `RecordingStopper.StopInstances` record every call:

- Add 1 to `r.Calls`.
- Append all of `params.InstanceIds` to `r.Stopped`.
- Then, if `r.Err` is set, return `nil` and `r.Err`. Otherwise return
  `&ec2.StopInstancesOutput{}` and `nil`.

A failing call is still recorded.

```go
r := &RecordingStopper{}
r.StopInstances(ctx, &ec2.StopInstancesInput{InstanceIds: []string{"i-web-1", "i-web-2"}})
r.StopInstances(ctx, &ec2.StopInstancesInput{InstanceIds: []string{"i-db-1"}})
// r.Calls   => 2
// r.Stopped => ["i-web-1" "i-web-2" "i-db-1"]
```

### 2. Stop `TagValue` from panicking

`TagValue(tags, key)` returns the value of the tag named `key`, or `""` if
there is none. It reads `*t.Key` and `*t.Value`, and panics when either is
`nil`. Use `aws.ToString` instead, so a `nil` key never matches and a `nil`
value reads as `""`.

```go
tags := []ec2types.Tag{{Key: aws.String("Env"), Value: aws.String("prod")}, {Key: aws.String("Owner")}}
TagValue(tags, "Env")   // => "prod"
TagValue(tags, "Owner") // => "" (Value is nil)
```

### 3. Fix `StopDevInstances`

`StopDevInstances(ctx, api, instances)` picks the instances to stop, sorts
their IDs and stops them in one call. It has two bugs:

- It stops every instance tagged `Env=dev`, even ones that are already
  stopped. Only pick instances whose `State` is not `nil` and whose
  `State.Name` is `ec2types.InstanceStateNameRunning`.
- When nothing needs stopping, it still calls `StopInstances` with an empty
  list. Return `nil, nil` instead, without calling the API.

The error wrapping (`stop instances: %w`) is already right. The tests pass in
your `RecordingStopper` from task 1, so finish that first.

```go
// i-dev-2 running Env=dev, i-prod-1 running Env=prod, i-dev-4 stopped Env=dev
StopDevInstances(ctx, spy, instances)
// => ["i-dev-2"], spy.Calls => 1

// i-prod-1 running Env=prod
StopDevInstances(ctx, spy, instances)
// => nil, spy.Calls => 0
```

## Learn more

- [pkg.go.dev: service/ec2 Client.StopInstances](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/ec2#Client.StopInstances)
- [Amazon EC2: Stop and start instances](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/Stop_Start.html)
- [pkg.go.dev: aws.ToString](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/aws#ToString)
