# Checksums and archives: publish a release

In *Checksums and archives: verify and extract safely* you consumed a release. Now you produce one: pack
the build output into a `.tar.gz` and publish a `SHA256SUMS` file next to it,
the way a CI release job does.

## Learn

**Recap:** `sha256.New()` is an `io.Writer`, so `io.Copy` streams a file into
it, and `hex.EncodeToString(h.Sum(nil))` gives the text hash. A `.tar.gz` is a
`tar.Writer` wrapped around a `gzip.Writer` wrapped around the output. Each
checksums line is `<hex>  <name>`, with two spaces.

### Close order matters

Both writers hold back data until they are closed. `tw.Close()` writes tar's
*end-of-archive marker* (two blocks of zero bytes) **into gzip**. `gz.Close()`
then flushes the compressed data and writes gzip's trailer. So the inner
writer (tar) must close first. If gzip closes first, the tar marker has
nowhere to go and the archive is truncated. Tools such as GNU `tar` reject it
with `Unexpected EOF in archive`.

Watch out for `defer`: deferred calls run in **reverse** order, last deferred
first, and both of these ignore the error from `Close`:

```go
defer tw.Close() // runs second
defer gz.Close() // runs first: wrong order!
```

For writers, closing is part of producing the output, so close them
explicitly and check the errors:

```go
if err := tw.Close(); err != nil {
	return err
}
return gz.Close()
```

### Listing a directory, sorted

`os.ReadDir(dir)` returns the entries directly inside `dir` (it doesn't
recurse) as a `[]os.DirEntry`, **already sorted by file name**. That gives
stable output, so CI logs don't reshuffle between runs.

```go
entries, err := os.ReadDir("dist")
for _, e := range entries {
	fmt.Println(e.Name(), e.IsDir(), e.Type().IsRegular())
}
```

`e.Type().IsRegular()` is true only for plain files: not directories, not
symlinks.

## Your task

Your release job builds into a source directory, then publishes to `dist/`.
Fix the packer and write the checksum step.

### 1. Fix the truncated archive

`Pack` writes every regular file under `dir` into a `.tar.gz`. Archives it
produces are missing the tar end-of-archive marker, because `gz.Close()` runs
before `tw.Close()`. Fix it: after the walk, close `tw`, then `gz`, and return
the first error. If the walk fails, return its error.

### 2. Write the checksums

Write `WriteChecksums(dir, w)`. For each entry from `os.ReadDir(dir)`, in the
order it returns them:

- Skip anything that isn't a regular file (subdirectories are not included).
- Skip a file named `SHA256SUMS` (the `ChecksumsFile` constant), so a re-run
  doesn't hash the old checksums.
- Hash the file and write `<hex>  <name>\n` to `w` with `fmt.Fprintf`.

Put the hashing in a small helper, `hashFile(path) (string, error)`, that
returns the SHA-256 of the file as lowercase hex. It's the same as
`SHA256File` from *Checksums and archives: verify and extract safely*: open,
`io.Copy` into `sha256.New()`, then `hex.EncodeToString`.

Return any error from `os.ReadDir`, hashing or writing. An empty directory
writes nothing.

```text
dist/
├── SHA256SUMS            (skipped)
├── agent-linux-amd64
├── notes/                (skipped)
└── web.tar.gz

=> 3f2a…  agent-linux-amd64
   9c41…  web.tar.gz
```

### 3. Bundle a release

Write `Bundle(srcDir, dist, name)`:

1. Create `dist` (and any parents) with `os.MkdirAll(dist, 0o755)`.
2. Create `dist/<name>.tar.gz` with `os.Create` and `Pack` `srcDir` into it.
   Close the file and check the error.
3. Create `dist/SHA256SUMS` with `os.Create` (this replaces an old one) and
   `WriteChecksums(dist, file)` into it. Close it and check the error.

Return the first error unchanged. For example, a missing `srcDir` makes `Pack`
fail with an error that `errors.Is(err, fs.ErrNotExist)` recognises.

```go
Bundle("build", "dist", "web-v2.0.0")
// dist/web-v2.0.0.tar.gz
// dist/SHA256SUMS lists web-v2.0.0.tar.gz and any other files already in dist
```

## Learn more

- [archive/tar: Writer.Close](https://pkg.go.dev/archive/tar#Writer.Close)
- [compress/gzip: Writer.Close](https://pkg.go.dev/compress/gzip#Writer.Close)
- [os.ReadDir](https://pkg.go.dev/os#ReadDir)
- [Go spec: Defer statements](https://go.dev/ref/spec#Defer_statements)
