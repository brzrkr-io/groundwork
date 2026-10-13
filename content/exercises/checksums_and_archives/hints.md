## Task 1: a hash is a writer

Open the file with `os.Open` and `defer f.Close()`. Then
`io.Copy(h, f)` with `h := sha256.New()`, and finish with
`hex.EncodeToString(h.Sum(nil))`.

## Task 2: find the line, then compare

```go
name := filepath.Base(filePath)
want := ""
scanner := bufio.NewScanner(sums)
for scanner.Scan() {
	fields := strings.Fields(scanner.Text())
	if len(fields) == 2 && fields[1] == name {
		want = fields[0]
		break
	}
}
```

If `want` is still `""`, return `fmt.Errorf("%w for %s", ErrNoChecksum, name)`.
Otherwise call `SHA256File(filePath)` and compare.

## Task 3: check the name first

The very first thing `extractEntry` does is:

```go
if !filepath.IsLocal(hdr.Name) {
	return fmt.Errorf("%w: %q", ErrUnsafePath, hdr.Name)
}
```

## Task 3: switch on the type

```go
target := filepath.Join(dst, hdr.Name)
switch hdr.Typeflag {
case tar.TypeDir:
	return os.MkdirAll(target, 0o755)
case tar.TypeReg:
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	return writeFile(target, tr, hdr.FileInfo().Mode().Perm())
default:
	return nil
}
```
