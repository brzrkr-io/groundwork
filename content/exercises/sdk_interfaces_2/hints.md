## Task 1: record first, then fail

Two lines record the call: `r.Calls++` and
`r.Stopped = append(r.Stopped, params.InstanceIds...)`. The `...` appends
every element of the slice. Only after that, check `r.Err`.

## Task 2: no stars

Replace `*t.Key` with `aws.ToString(t.Key)` and `*t.Value` with
`aws.ToString(t.Value)`. Add `"github.com/aws/aws-sdk-go-v2/aws"` to the
imports.

## Task 3: skip, then guard

At the top of the loop, skip instances that aren't running:

```go
if inst.State == nil || inst.State.Name != ec2types.InstanceStateNameRunning {
	continue
}
```

After sorting, return early when there is nothing to stop:

```go
if len(ids) == 0 {
	return nil, nil
}
```
