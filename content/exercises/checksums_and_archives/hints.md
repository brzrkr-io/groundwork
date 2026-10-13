## Task 1: Hash a file

A hash is an `io.Writer`: whatever you write into it gets hashed. Open the
file and copy it straight into the hash.

### Closer

Open with `os.Open` and `defer f.Close()`. Make the hash with
`h := sha256.New()`, then `io.Copy(h, f)`. Finish with
`hex.EncodeToString(h.Sum(nil))`.

### Answer

```go
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
```

## Task 2: Verify against a checksums file

First find the expected hash: scan the file line by line for the one whose
name matches. Only then hash the file and compare.

### Closer

```go
scanner := bufio.NewScanner(sums)
for scanner.Scan() {
	fields := strings.Fields(scanner.Text())
	// two fields, and fields[1] == name?
}
```

If no line matched, return `fmt.Errorf("%w for %s", ErrNoChecksum, name)`.

### Answer

```go
sums, err := os.Open(sumsPath)
if err != nil {
	return err
}
defer sums.Close()

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
if err := scanner.Err(); err != nil {
	return err
}
if want == "" {
	return fmt.Errorf("%w for %s", ErrNoChecksum, name)
}
got, err := SHA256File(filePath)
if err != nil {
	return err
}
if got != want {
	return fmt.Errorf("%w for %s: got %s, want %s", ErrChecksumMismatch, name, got, want)
}
return nil
```

## Task 3: Extract safely

Check the entry's name before anything else, so nothing outside `dst` is ever
touched. Then decide what to do from the entry's type.

### Closer

```go
if !filepath.IsLocal(hdr.Name) {
	return fmt.Errorf("%w: %q", ErrUnsafePath, hdr.Name)
}
target := filepath.Join(dst, hdr.Name)
switch hdr.Typeflag {
case tar.TypeDir:
	// ...
}
```

### Answer

```go
if !filepath.IsLocal(hdr.Name) {
	return fmt.Errorf("%w: %q", ErrUnsafePath, hdr.Name)
}
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
