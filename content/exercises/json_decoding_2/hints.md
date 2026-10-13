## Task 1: look at the second argument

`json.Unmarshal` has to change `list`. What do you pass when a function must
change your variable? Add one character.

## Task 2: exported and exact

The json package can't see lowercase fields: rename `namespace` to
`Namespace` (in the struct **and** in `Ref`). Then change the tag on `Phase`
to `json:"phase"`.

## Task 3: the loop shape

Create the decoder once, before the loop. Keep a counter `n` that goes up by
one per object. Inside the loop: declare `var pod Pod`, call
`dec.Decode(&pod)`, stop cleanly on `io.EOF`, return on any other error, and
otherwise do `counts[pod.Status.Phase]++`.

## Almost the answer

```go
dec := json.NewDecoder(r)
for n := 1; ; n++ {
	var pod Pod
	err := dec.Decode(&pod)
	if errors.Is(err, io.EOF) {
		return counts, nil
	}
	if err != nil {
		return nil, fmt.Errorf("pod %d: %w", n, err)
	}
	counts[pod.Status.Phase]++
}
```
