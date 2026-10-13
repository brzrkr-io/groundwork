## Task 1: Finish the spy

Record the call first, then decide whether to fail. That way a failing call
is still counted and its IDs are still saved.

### Closer

`r.Calls++` counts the call. `append(r.Stopped, params.InstanceIds...)` adds
every element of the slice; the `...` spreads it out. Only after that, check
`r.Err`.

### Answer

```go
r.Calls++
r.Stopped = append(r.Stopped, params.InstanceIds...)
if r.Err != nil {
	return nil, r.Err
}
return &ec2.StopInstancesOutput{}, nil
```

## Task 2: Stop `TagValue` from panicking

The stars in `*t.Key` and `*t.Value` are the problem: they panic on a `nil`
pointer. There is an `aws` helper that reads a `*string` safely.

### Closer

`aws.ToString(p)` returns `""` when `p` is `nil`. Add the import
`"github.com/aws/aws-sdk-go-v2/aws"` and use it in place of each `*`.

### Answer

```go
for _, t := range tags {
	if aws.ToString(t.Key) == key {
		return aws.ToString(t.Value)
	}
}
return ""
```

## Task 3: Fix `StopDevInstances`

Two fixes: skip instances that aren't running, and don't call the API when
there is nothing to stop.

### Closer

At the top of the loop, `continue` when `inst.State` is `nil` or its `Name` is
not `ec2types.InstanceStateNameRunning`. After sorting, check `len(ids)`
before calling `StopInstances`.

### Answer

```go
for _, inst := range instances {
	if inst.State == nil || inst.State.Name != ec2types.InstanceStateNameRunning {
		continue
	}
	if TagValue(inst.Tags, "Env") == "dev" {
		ids = append(ids, aws.ToString(inst.InstanceId))
	}
}
slices.Sort(ids)

if len(ids) == 0 {
	return nil, nil
}
```

AWS rejects a `StopInstances` call with no IDs, so returning early skips it.
