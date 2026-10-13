## Task 1: stop deferring the closes

Delete both `defer` lines. Store the result of `filepath.WalkDir` in `err`
instead of returning it straight away, return it if it isn't `nil`, then
close `tw` and finally return `gz.Close()`.

## Task 2: the hashing helper

```go
func hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
```

## Task 2: the loop

```go
entries, err := os.ReadDir(dir)
if err != nil {
	return err
}
for _, e := range entries {
	if !e.Type().IsRegular() || e.Name() == ChecksumsFile {
		continue
	}
	// hash filepath.Join(dir, e.Name()) and Fprintf the line
}
```

## Task 2: the line format

```go
fmt.Fprintf(w, "%s  %s\n", sum, e.Name())
```

## Task 3: one file at a time

```go
archive, err := os.Create(filepath.Join(dist, name+".tar.gz"))
if err != nil {
	return err
}
if err := Pack(srcDir, archive); err != nil {
	archive.Close()
	return err
}
if err := archive.Close(); err != nil {
	return err
}
```

Repeat the same shape for `SHA256SUMS` with `WriteChecksums(dist, sums)`.
