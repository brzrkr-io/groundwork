## Task 1: Add deploys to the queue

A newcomer joins at the back. One `list` method adds there and already
returns the element you need to give back.

### Closer

`q.deploys.PushBack(d)` adds `d` at the back and returns its
`*list.Element`. Return that.

### Answer

```go
return q.deploys.PushBack(d)
```

## Task 2: List the queue

Make a slice, then walk the list from the front, following `Next()` until
it's `nil`. Each element's `Value` holds a `*Deploy`.

### Closer

`for el := q.deploys.Front(); el != nil; el = el.Next() { ... }`. Inside,
`el.Value.(*Deploy).Service` is the service to `append`.

### Answer

```go
services := make([]string, 0, q.deploys.Len())
for el := q.deploys.Front(); el != nil; el = el.Next() {
	services = append(services, el.Value.(*Deploy).Service)
}
return services
```

## Task 3: Let a hotfix jump the line

You already hold the element. The list has a method that moves an element
to the front without removing it.

### Closer

`q.deploys.MoveToFront(...)`.

### Answer

```go
q.deploys.MoveToFront(el)
```

## Task 4: Take the next deploy

Get the front element and handle the empty case first. Then remove it: the
value `Remove` hands back is the deploy, as `any`.

### Closer

`front := q.deploys.Front()` is `nil` on an empty queue.
`q.deploys.Remove(front)` returns `any`; assert it with `.(*Deploy)`.

### Answer

```go
front := q.deploys.Front()
if front == nil {
	return nil, false
}
return q.deploys.Remove(front).(*Deploy), true
```
