# Checksums and archives: publish a release

In *Checksums and archives: verify and extract safely* you consumed a
release. Now you produce one: pack the build output into a `.tar.gz` and
publish a `SHA256SUMS` file next to it, the way a CI release job does.

**New here:** close order of stacked writers, `os.ReadDir`. (`defer` order is from *defer*, `os.MkdirAll` from *Files: an append-only audit log*.)

## Learn

### Recap

- `sha256.New()` is an `io.Writer`, so `io.Copy` streams a file into it.
- `hex.EncodeToString(h.Sum(nil))` turns the hash into text.
- A `.tar.gz` is a `tar.Writer` wrapped around a `gzip.Writer` wrapped around the output.
- A checksums line is `<hex>  <name>`, with **two** spaces.

### Close order matters

Both archive writers hold back data until they're closed:

- Closing tar writes its *end-of-archive marker* (two blocks of zero bytes) **into gzip**.
- Closing gzip flushes the compressed data and writes gzip's trailer.

So the inner writer (tar) must close first. If gzip closes first, tar's
marker has nowhere to go and the archive is truncated. GNU `tar` rejects it
with `Unexpected EOF in archive`.

### `defer` runs in reverse

Deferred calls run last-deferred-first, and a deferred `Close` throws its
error away:

```go
defer fmt.Println("one") // prints second
defer fmt.Println("two") // prints first
```

For a writer, closing is part of producing the output. So close writers
explicitly, in the right order, and check each error. The same pattern
appears with a buffered writer around a file:

```go
if err := bw.Flush(); err != nil { // inner writer first
	return err
}
return f.Close() // outer last; its error is the result
```

> **Why is it written like this?** `if err := ...; err != nil` declares `err`
> only inside the `if`. Go allows that short form; using it to keep `err`
> local is convention. Returning the last `Close` directly is the author's
> choice; it saves an `if`.

### Listing a directory, sorted

`os.ReadDir(dir)` returns the entries directly inside `dir` (it doesn't go
into subdirectories) as a `[]os.DirEntry`, **already sorted by name**. Stable
order means CI logs don't reshuffle between runs.

```go
entries, err := os.ReadDir("/etc/nginx")
for _, e := range entries {
	fmt.Println(e.Name(), e.IsDir(), e.Type().IsRegular())
}
```

`e.Name()` is just the file name, not the full path. `e.Type().IsRegular()`
is true only for plain files: not directories, not symlinks.

### Making directories

`os.MkdirAll(path, 0o755)` creates `path` and any missing parents, and does
nothing if it already exists. `0o755` is the permission: owner can write,
everyone can read and enter.

> **Why is it written like this?** `0o` marks an *octal* (base-8) number.
> Go requires a prefix for octal; `0o` is the modern style (a bare leading
> `0` also works but is easy to misread).

## Your task

Your release job builds into a source directory, then publishes to `dist/`.
Fix the packer and write the checksum step.

### 1. Fix the truncated archive

**Goal:** make `Pack` close `tw` and then `gz`, returning any error, so the archive is complete.

- `Pack` writes every regular file under `dir` into a `.tar.gz`. That part works.
- Right now the two `defer` lines close `gz` before `tw`, and ignore errors.
- If the walk fails, return its error.
- Otherwise close `tw`, then `gz`, and return the first error.

**Think it through:**
- Right now the walk's result is returned at once. Where must it go instead so code can run after it?
- Which writer writes into the other, so which must close first?

### 2. Write the checksums

**Goal:** write `WriteChecksums(dir, w)` so it writes one `<hex>  <name>\n` line to `w` per regular file in `dir`.

- Use the entries from `os.ReadDir(dir)`, in the order it returns them.
- Skip anything that isn't a regular file (subdirectories are left out).
- Skip the file named by the `ChecksumsFile` constant (`SHA256SUMS`), so a re-run doesn't hash old checksums.
- Write each line with `fmt.Fprintf`. `<name>` is the file name only; there are two spaces before it.
- Return any error from `os.ReadDir`, hashing or writing. An empty directory writes nothing.
- Add `crypto/sha256`, `encoding/hex` and `fmt` to the imports.

Put the hashing in a helper, `hashFile(path string) (string, error)`, that
returns the file's SHA-256 as lowercase hex. It's the same as `SHA256File`
from *Checksums and archives: verify and extract safely*.

```text
dist/
├── SHA256SUMS            (skipped)
├── agent-linux-amd64
├── notes/                (skipped)
└── web.tar.gz

=> 3f2a…  agent-linux-amd64
   9c41…  web.tar.gz
```

**Think it through:**
- `e.Name()` is only the name. What do you need to join it with to open the file?
- Which two conditions mean "skip this entry"?

### 3. Bundle a release

**Goal:** write `Bundle(srcDir, dist, name)` so it creates `dist/<name>.tar.gz` and a fresh `dist/SHA256SUMS`.

In order:

1. Create `dist` (and any parents) with `os.MkdirAll`, permission `0o755`.
2. Create `dist/<name>.tar.gz` with `os.Create` and `Pack` `srcDir` into it. Close the file and check the error.
3. Create `dist/SHA256SUMS` with `os.Create` (this replaces an old one) and `WriteChecksums(dist, ...)` into it. Close it and check the error.

Return the first error unchanged. For example, a missing `srcDir` makes
`Pack` fail with an error that `errors.Is(err, fs.ErrNotExist)` recognises.

```go
Bundle("build", "dist", "web-v2.0.0")
// creates dist/web-v2.0.0.tar.gz
// dist/SHA256SUMS lists web-v2.0.0.tar.gz and any other files already in dist
```

**Think it through:**
- The archive name is built from the `name` parameter plus fixed text. Which part goes in quotes?
- Use `filepath.Join` to put `dist` in front of a file name. Use the `ChecksumsFile` constant rather than retyping it.
- If `Pack` fails, the file is still open. What should happen to it before you return?

## Learn more

- [archive/tar: Writer.Close](https://pkg.go.dev/archive/tar#Writer.Close)
- [compress/gzip: Writer.Close](https://pkg.go.dev/compress/gzip#Writer.Close)
- [os.ReadDir](https://pkg.go.dev/os#ReadDir)
- [Go spec: Defer statements](https://go.dev/ref/spec#Defer_statements)
