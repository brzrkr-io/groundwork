## Task 1: Pick a shard for a key

Three steps: make an FNV-1a hasher, write the key into it, and take the
32-bit sum modulo `n`. Remember to add `"hash/fnv"` to the imports.

### Closer

`h := fnv.New32a()`, then `h.Write([]byte(key))`. The modulo has to be
between two `uint32`s: `h.Sum32() % uint32(n)`. Wrap the whole thing in
`int(...)` to return it.

### Answer

```go
h := fnv.New32a()
h.Write([]byte(key))
return int(h.Sum32() % uint32(n))
```

## Task 2: Group keys by shard

Make the outer slice with `n` elements first. Then loop over the keys, ask
`ShardFor` for each key's shard, and append the key to that shard's slice.

### Closer

`shards := make([][]string, n)` gives `n` nil slices. Inside the loop,
`i := ShardFor(key, n)`, then append to `shards[i]` and store the result
back in `shards[i]`.

### Answer

```go
shards := make([][]string, n)
for _, key := range keys {
	i := ShardFor(key, n)
	shards[i] = append(shards[i], key)
}
return shards
```

## Task 3: Plan a rebalance

For each key, work out its shard with `from` shards and with `to` shards.
If the two differ, the key moves: append a `Move` to the result.

### Closer

`before, after := ShardFor(key, from), ShardFor(key, to)`. Start with
`var moves []Move`, and append
`Move{Key: key, From: before, To: after}` when `before != after`.

### Answer

```go
var moves []Move
for _, key := range keys {
	before, after := ShardFor(key, from), ShardFor(key, to)
	if before != after {
		moves = append(moves, Move{Key: key, From: before, To: after})
	}
}
return moves
```

## Task 4: Measure how much moves

Handle the empty case first, so you never divide by zero. Then divide the
length of `Plan`'s result by the number of keys.

### Closer

`if len(keys) == 0 { return 0 }`. Convert both lengths with `float64(...)`
before dividing.

### Answer

```go
if len(keys) == 0 {
	return 0
}
return float64(len(Plan(keys, from, to))) / float64(len(keys))
```
