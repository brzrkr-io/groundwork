## Task 1: Cache only successful reads

The store's answer only belongs in the cache once you know it's real. Check
the error first, and only cache after that check has passed.

### Closer

Move `s.cache.Put(name, value)` below the `if err != nil { return "", err }`
block, so the error path returns before anything is cached.

### Answer

```go
value, err := s.store.Get(name)
if err != nil {
	return "", err
}
s.cache.Put(name, value)
return value, nil
```

## Task 2: Write the store first, then invalidate

Two changes: the store write comes first and can stop everything, and the
cache step removes the key instead of putting the new value in.

### Closer

Start with `if err := s.store.Set(name, value); err != nil { return err }`.
After that, call `s.cache.Remove(name)` and return `nil`. Delete the
`s.cache.Put` line.

### Answer

```go
if err := s.store.Set(name, value); err != nil {
	return err
}
s.cache.Remove(name)
return nil
```

## Task 3: Invalidate on delete

Same shape as `Set`: change the store, stop on an error, then remove the
cached copy.

### Closer

Check the error from `s.store.Delete(name)` with `if err := ...; err != nil`,
then call `s.cache.Remove(name)` and return `nil`.

### Answer

```go
if err := s.store.Delete(name); err != nil {
	return err
}
s.cache.Remove(name)
return nil
```
