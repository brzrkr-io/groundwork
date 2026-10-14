## Task 1: Load the trail

Two steps, each with an error check: read the file's bytes, then decode them
into something with a `Records` field.

### Closer

`os.ReadFile(path)` gives `[]byte`. Decode into an anonymous struct:

```go
var file struct {
	Records []Event `json:"Records"`
}
```

then `json.Unmarshal(data, &file)` and return `file.Records`.

### Answer

```go
data, err := os.ReadFile(path)
if err != nil {
	return nil, err
}
var file struct {
	Records []Event `json:"Records"`
}
if err := json.Unmarshal(data, &file); err != nil {
	return nil, fmt.Errorf("%s: %w", path, err)
}
return file.Records, nil
```

## Task 2: Follow one access key

Filter first, into a new slice, then sort only that new slice. The key ID
lives inside `UserIdentity`.

### Closer

`if e.UserIdentity.AccessKeyID == keyID { found = append(found, e) }`, then
`slices.SortFunc(found, func(a, b Event) int { ... })` using
`a.EventTime.Compare(b.EventTime)`.

### Answer

```go
var found []Event
for _, e := range events {
	if e.UserIdentity.AccessKeyID == keyID {
		found = append(found, e)
	}
}
slices.SortFunc(found, func(a, b Event) int {
	return a.EventTime.Compare(b.EventTime)
})
return found
```

## Task 3: Summarize the incident

One loop over the events does everything: track the earliest and latest time,
put IPs, regions and actions into three sets, and count errors. After the
loop, turn each set into a sorted slice. Handle "no events" first.

### Closer

Sets are `map[string]bool{}`; add with `ips[e.SourceIPAddress] = true`.
`slices.Sorted(maps.Keys(ips))` gives the sorted, unique list.
`strings.TrimSuffix(e.EventSource, ".amazonaws.com")` gives the service.
Start `First` and `Last` at `events[0].EventTime`, then update them with
`Before` and `After`.

### Answer

```go
var s Summary
if len(events) == 0 {
	return s
}
ips, regions, actions := map[string]bool{}, map[string]bool{}, map[string]bool{}
s.First, s.Last = events[0].EventTime, events[0].EventTime
for _, e := range events {
	if e.EventTime.Before(s.First) {
		s.First = e.EventTime
	}
	if e.EventTime.After(s.Last) {
		s.Last = e.EventTime
	}
	ips[e.SourceIPAddress] = true
	regions[e.AWSRegion] = true
	service := strings.TrimSuffix(e.EventSource, ".amazonaws.com")
	actions[service+":"+e.EventName] = true
	if e.ErrorCode != "" {
		s.Denied++
	}
}
s.SourceIPs = slices.Sorted(maps.Keys(ips))
s.Regions = slices.Sorted(maps.Keys(regions))
s.Actions = slices.Sorted(maps.Keys(actions))
return s
```
