# Checksums and archives: verify and extract safely

Install scripts download a release, check it hasn't been tampered with, and
unpack it. You'll build those three steps, including the check that stops a
malicious archive from writing files outside the install folder.

**New here:** `crypto/sha256`, `encoding/hex`, `archive/tar`, `compress/gzip`, `filepath.IsLocal`.

## Learn

### Checksums

A *checksum* is a short fingerprint of a file's bytes. Change one byte and
the fingerprint changes completely. Releases publish a `SHA256SUMS` file in
the format the `sha256sum` tool prints: 64 hex characters, **two spaces**,
the file name:

```text
5891b5b522d5df086d0ff0b110fbd9d21bb4fc7163af34d08286a2e846f6be03  app.tar.gz
```

### Hashing in Go

A hash (`hash.Hash`) is an `io.Writer`: every byte written into it feeds the
fingerprint. So `io.Copy` can stream any reader into it, a chunk at a time,
without loading a 2 GB file into memory. Here with MD5 and a string, to show
the shape:

```go
h := md5.New()
io.Copy(h, strings.NewReader("hi"))
raw := h.Sum(nil)              // the finished hash as raw bytes
text := hex.EncodeToString(raw) // "49f68a5c8493ec2c0bf489821c21fc3b"
```

`sha256.New()` (package `crypto/sha256`) works the same way.
`hex.EncodeToString` (package `encoding/hex`) gives lowercase hex.

> **Why is it written like this?** `h.Sum(nil)` takes a slice to append the
> hash to; `nil` means "start a fresh one". That's the API's design, and
> `nil` is what almost everyone passes.

### tar and gzip

A `.tar.gz` is two formats stacked. *tar* bundles files: each entry has a
*header* (name, size, mode, type) then its contents. *gzip* compresses the
whole stream. In Go each layer wraps the one below:

```go
gz, err := gzip.NewReader(r) // r is the compressed bytes
tr := tar.NewReader(gz)      // reads entries out of the gzip stream
hdr, err := tr.Next()        // move to the next entry; io.EOF = no more
// hdr.Name is its path; reading from tr reads its contents
```

`Extract` in `exercise.go` already has this loop. `Create` is a complete,
commented example of *writing* a `.tar.gz`; the tests use it.

`hdr.Typeflag` says what an entry is:

| Typeflag | Meaning |
|---|---|
| `tar.TypeDir` | a directory |
| `tar.TypeReg` | a regular file |
| `tar.TypeSymlink` and others | links and special files |

`hdr.FileInfo().Mode().Perm()` gives the permission bits, such as `0o755`.

### Zip slip

An entry's name is just a string the archive's author chose. A malicious one
might be `../../etc/cron.d/evil` or `/etc/passwd`. Join that onto your target
folder and you write outside it. This attack is called *zip slip*.

`filepath.IsLocal(name)` reports whether a path stays inside the folder it's
joined to (relative, not empty, no climbing out with `..`):

| Name | `filepath.IsLocal` |
|---|---|
| `bin/agent` | `true` |
| `logs/../app.yaml` | `true` |
| `../evil.txt` | `false` |
| `/abs.txt` | `false` |

Check every name **before** creating anything.

## Your task

Your installer downloads `app.tar.gz` and `SHA256SUMS`. Make it hash, verify
and safely unpack the release.

### 1. Hash a file

**Goal:** write `SHA256File(path)` to return the file's SHA-256 as lowercase
hex.

Open the file, stream it into `sha256.New()` with `io.Copy`, return the hex
string and `nil`. If the file can't be opened (or read), return `""` and the
error. Add `"crypto/sha256"` and `"encoding/hex"` to the imports.

```go
SHA256File("hello.txt") // file contains "hello\n"
// => "5891b5b522d5df086d0ff0b110fbd9d21bb4fc7163af34d08286a2e846f6be03", nil
```

An empty file hashes to `"e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"`.

**Think it through:**
- Which value do you open: the parameter `path`, or the literal word `"path"`?
- Close the file when done. Which keyword runs that at the end?

### 2. Verify against a checksums file

**Goal:** write `Verify(sumsPath, filePath)`: return `nil` when the file's
hash matches its line in the checksums file.

- Read `sumsPath` line by line with `bufio.Scanner`.
- Split each line with `strings.Fields`. A valid line has 2 fields: hash, name.
- The right line's name equals `filepath.Base(filePath)` **exactly**
  (`app.tar` must not match `app.tar.gz`).

| Situation | Return (exact text) |
|---|---|
| hash matches | `nil` |
| no line for the name | wraps `ErrNoChecksum`: `no checksum found for app.tar.gz` |
| hash differs | wraps `ErrChecksumMismatch`: `checksum mismatch for app.tar.gz: got <hex>, want <hex>` |

`got` is the file's actual hash; `want` is the one from the checksums file.
Build the messages with `%w` first, then the rest, e.g. `"%w for %s"`.
Errors from opening, reading (`scanner.Err()` after the loop) or hashing are
returned unchanged. Add `"bufio"`, `"fmt"` and `"strings"` to the imports.

**Think it through:**
- Find the expected hash first. Only then hash the file and compare.
- After the loop, how do you know no line matched?

### 3. Extract safely

**Goal:** rewrite `extractEntry` so it rejects unsafe names, creates
directories, writes regular files and skips everything else.

Right now it writes every entry as a file wherever its name points. Do this,
in order:

1. If `hdr.Name` isn't local, return an error wrapping `ErrUnsafePath` with
   the name in quotes: `unsafe path in archive: "../evil.txt"` (format
   `%w: %q`).
2. The target path is `dst` joined with `hdr.Name`.
3. `tar.TypeDir`: create the target folder with `os.MkdirAll(…, 0o755)`.
4. `tar.TypeReg`: create the target's **parent** folder, then call
   `writeFile` with the target, `tr` and the entry's permission bits.
5. Any other type (symlinks, devices): do nothing, return `nil`.

```go
Extract(archive, "/opt/app")
// bin/web          => /opt/app/bin/web, executable
// tmp/cache/       => an empty folder /opt/app/tmp/cache
// ../evil.txt      => error: unsafe path in archive: "../evil.txt", nothing written
```

**Think it through:**
- Step 1 must come before anything touches the disk. Why?
- `hdr.Typeflag` has three outcomes. Which Go statement picks between several values?
- `filepath.Dir` gives a path's parent folder.

## Learn more

- [crypto/sha256 package](https://pkg.go.dev/crypto/sha256)
- [archive/tar package](https://pkg.go.dev/archive/tar)
- [filepath.IsLocal](https://pkg.go.dev/path/filepath#IsLocal)
- [Go blog: Traversal-resistant file APIs](https://go.dev/blog/osroot)
