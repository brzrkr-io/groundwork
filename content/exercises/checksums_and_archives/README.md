# Checksums and archives: verify and extract safely

Install scripts download a release, check it hasn't been tampered with, and
unpack it. Here you build those three steps in Go, including the safety check
that stops a malicious archive from overwriting files elsewhere on a host.

## Learn

### Checksums

A *checksum* is a short fingerprint of a file's bytes. Change one byte and the
fingerprint changes completely. Releases publish a checksums file in the
format the `sha256sum` tool prints: 64 hex characters, **two spaces**, then
the file name:

```text
5891b5b522d5df086d0ff0b110fbd9d21bb4fc7163af34d08286a2e846f6be03  app.tar.gz
```

On the command line, `sha256sum -c SHA256SUMS` re-hashes each listed file and
reports `OK` or `FAILED`.

### Hashing a file in Go

`sha256.New()` (package `crypto/sha256`) returns a `hash.Hash`. A hash is an
`io.Writer`: every byte you write is fed into the fingerprint. So `io.Copy`
can stream a 2 GB file into it a chunk at a time, without loading the whole
file into memory:

```go
h := sha256.New()
if _, err := io.Copy(h, f); err != nil { // f is an *os.File
	return "", err
}
sum := h.Sum(nil)               // the raw 32 bytes
text := hex.EncodeToString(sum) // "5891b5b5...", lowercase hex
```

`h.Sum(nil)` returns the finished hash. `hex.EncodeToString` (package
`encoding/hex`) turns the bytes into the familiar text form.

### tar and gzip

A `.tar.gz` file is two formats stacked. *tar* bundles many files into one
stream: each file gets a *header* (name, size, mode, type), followed by its
contents. *gzip* compresses that stream. In Go each layer wraps the one
below, just as `bufio.Scanner` wraps a reader:

```go
gz, err := gzip.NewReader(r) // r is the compressed .tar.gz bytes
// handle err, then: defer gz.Close()
tr := tar.NewReader(gz)      // reads tar entries out of the gzip stream
for {
	hdr, err := tr.Next() // move to the next entry
	if err == io.EOF {
		break // no more entries
	}
	if err != nil {
		return err
	}
	// hdr.Name is the entry's path; reading from tr reads its contents.
}
```

`hdr.Typeflag` says what an entry is:

| Typeflag | Meaning |
|---|---|
| `tar.TypeDir` | a directory |
| `tar.TypeReg` | a regular file |
| `tar.TypeSymlink` and others | links and special files |

`hdr.FileInfo().Mode().Perm()` gives the file's permission bits, such as
`0o755` for an executable.

Read `Create` in `exercise.go`: it is a complete, commented example of
*writing* a `.tar.gz`. The tests use it to build archives for you to extract.

### Zip slip

An archive entry's name is just a string chosen by whoever made the archive.
A malicious one might be called `../../etc/cron.d/evil` or `/etc/passwd`.
Join that onto your target directory and you write outside it. This attack is
called *zip slip*.

`filepath.IsLocal(name)` reports whether a path stays inside the directory it
is joined to: it must be relative, must not be empty, and must not climb out
with `..`.

| Name | `filepath.IsLocal` |
|---|---|
| `bin/agent` | `true` |
| `logs/../app.yaml` | `true` |
| `../evil.txt` | `false` |
| `/abs.txt` | `false` |

Check every name **before** you create anything.

## Your task

Your team's installer downloads `app.tar.gz` and a `SHA256SUMS` file. Make it
hash, verify and safely unpack the release.

### 1. Hash a file

Write `SHA256File(path)`. Open the file, `io.Copy` it into `sha256.New()`, and
return the hash as lowercase hex. Return the error if the file can't be opened.

```go
SHA256File("hello.txt") // file contains "hello\n"
// => "5891b5b522d5df086d0ff0b110fbd9d21bb4fc7163af34d08286a2e846f6be03", nil
```

### 2. Verify against a checksums file

Write `Verify(sumsPath, filePath)`. Read the checksums file line by line with
`bufio.Scanner`. Split each line with `strings.Fields`; a valid line has two
fields, the hash and a name. Find the line whose name is exactly
`filepath.Base(filePath)` (`app.tar` must not match `app.tar.gz`).

| Situation | Return |
|---|---|
| Hash matches | `nil` |
| No line for the name | an error wrapping `ErrNoChecksum`: `no checksum found for app.tar.gz` |
| Hash differs | an error wrapping `ErrChecksumMismatch`: `checksum mismatch for app.tar.gz: got <hex>, want <hex>` |

`got` is the file's actual hash and `want` is the one from the checksums file.
Wrap with `fmt.Errorf("%w for %s", ErrNoChecksum, name)` and similar. Return
any error from opening or hashing as it is.

### 3. Extract safely

`Extract` already loops over the entries and calls `extractEntry` for each
one. Right now `extractEntry` writes every entry as a regular file, wherever
its name points: a directory entry fails, a symlink becomes a plain file, and
`../evil.txt` lands outside `dst`. Rewrite it to handle one entry like this:

1. If `filepath.IsLocal(hdr.Name)` is false, return an error wrapping
   `ErrUnsafePath` with the name quoted:
   `unsafe path in archive: "../evil.txt"` (use `%w: %q`).
2. Build the target path with `filepath.Join(dst, hdr.Name)`.
3. `tar.TypeDir`: create it with `os.MkdirAll(target, 0o755)`.
4. `tar.TypeReg`: create its parent directory with `os.MkdirAll`, then call
   `writeFile(target, tr, hdr.FileInfo().Mode().Perm())`.
5. Any other type (symlinks, devices): skip it and return `nil`.

```go
Extract(archive, "/opt/app")
// bin/web          => /opt/app/bin/web, mode 0755
// ../evil.txt      => error: unsafe path in archive: "../evil.txt"
```

## Learn more

- [crypto/sha256 package](https://pkg.go.dev/crypto/sha256)
- [archive/tar package](https://pkg.go.dev/archive/tar)
- [filepath.IsLocal](https://pkg.go.dev/path/filepath#IsLocal)
- [Go blog: Traversal-resistant file APIs](https://go.dev/blog/osroot)
