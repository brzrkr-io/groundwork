## Task 1: Fix the truncated archive

The two `defer` lines close the writers in the wrong order and ignore their
errors. tar must close first, because it writes its end-of-archive marker into
gzip.

### Closer

Delete both `defer` lines. Keep the result of `filepath.WalkDir` in `err`
instead of returning it straight away, return it if it isn't `nil`, then close
`tw` and return `gz.Close()`.

### Answer

```go
err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
	// ... unchanged ...
})
if err != nil {
	return err
}
if err := tw.Close(); err != nil {
	return err
}
return gz.Close()
```

## Task 2: Write the checksums

Start with the `hashFile` helper: open the file, stream it into
`sha256.New()`, and turn the sum into hex. Then loop over `os.ReadDir(dir)`
and skip anything that isn't a regular file or is named `ChecksumsFile`.

### Closer

`io.Copy(h, f)` hashes the file and `hex.EncodeToString(h.Sum(nil))` gives the
text. In the loop, `!e.Type().IsRegular() || e.Name() == ChecksumsFile` means
`continue`. The line is `fmt.Fprintf(w, "%s  %s\n", sum, e.Name())`, with two
spaces.

### Answer

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

func WriteChecksums(dir string, w io.Writer) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if !e.Type().IsRegular() || e.Name() == ChecksumsFile {
			continue
		}
		sum, err := hashFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "%s  %s\n", sum, e.Name()); err != nil {
			return err
		}
	}
	return nil
}
```

## Task 3: Bundle a release

Do the three steps one at a time: make the directory, write the archive, then
write the checksums. Each file you create must be closed, and the error from
`Close` checked.

### Closer

The archive step has this shape; repeat it for `SHA256SUMS` with
`WriteChecksums(dist, sums)`:

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

### Answer

```go
if err := os.MkdirAll(dist, 0o755); err != nil {
	return err
}
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
sums, err := os.Create(filepath.Join(dist, ChecksumsFile))
if err != nil {
	return err
}
if err := WriteChecksums(dist, sums); err != nil {
	sums.Close()
	return err
}
return sums.Close()
```
