# AWS SDK configuration: regions, profiles and credentials

In *Cloud fundamentals*, ending with *Practice: review an AWS account*, you
learned where AWS resources live and who may touch them. Now you call AWS
from Go. Every such program starts the same way: pick *which region* to talk
to and *which credentials* to sign requests with, without ever printing the
secret.

**New here:** the AWS SDK for Go, `aws.Config`, `config.LoadDefaultConfig`, functional options, credentials providers, profiles and the shared config files.

## Learn

### Recap: regions and credentials

From the guides *The cloud from the ground up* and
*Cloud and cluster access: IAM, RBAC and short-lived credentials*:

- Most resources live in one *region* (`us-east-1`, `eu-west-1`). An API
  request goes to one region and only sees what is there, so every client
  needs a region.
- Every request is signed with *credentials*, and IAM checks them against
  the policies, as in *IAM policies: how AWS decides allow or deny*.
- Credentials are either long-term user keys or temporary role credentials,
  as in *Assume role: trust policies and temporary credentials*:

| Part | Example | Safe to log? |
|---|---|---|
| *access key ID* | `AKIA...` (long-term) or `ASIA...` (temporary) | only shortened |
| *secret access key* | `wJal...` | never |
| *session token* | only for temporary credentials | never |

Programs running in AWS should get temporary credentials from a role. On a
laptop, and in this exercise, they come from files.

### Profiles and the two shared files

The `aws` CLI and the Go SDK read two files. A *profile* is a named section,
so one laptop can hold settings for several accounts:

```ini
# ~/.aws/config: settings. Non-default sections say "profile <name>".
[default]
region = us-east-1

[profile staging]
region = eu-west-1
```

```ini
# ~/.aws/credentials: keys. Sections are just "<name>".
[staging]
aws_access_key_id = AKIASTAGING000EXAMPLE
aws_secret_access_key = stagingSecretKey/K7MDENG/EXAMPLEKEY
```

### Loading an aws.Config

Every SDK client is built from an `aws.Config` (package `aws`). It holds the
`Region` and a `Credentials` provider. `config.LoadDefaultConfig` (package
`config`) builds one from your environment and the shared files.

You change its behaviour with *functional options*: small values that each
change one setting. You pass as many as you like after `ctx`:

```go
cfg, err := config.LoadDefaultConfig(ctx,
	config.WithRegion("ap-south-1"),
	config.WithSharedConfigProfile("billing"),
)
```

> **Why is it written like this?** Go requires the commas between arguments,
> including the one after the last argument when the `)` is on its own line.
> One option per line is convention (gofmt keeps it readable). Which options
> you pass, and in which order, is your choice: order doesn't matter here.

| Option | Effect |
|---|---|
| `config.WithRegion(r)` | use region `r` |
| `config.WithSharedConfigProfile(p)` | use profile `p` instead of `default` |
| `config.WithSharedConfigFiles([]string{path})` | read this config file instead of `~/.aws/config` |
| `config.WithSharedCredentialsFiles([]string{path})` | read this credentials file instead of `~/.aws/credentials` |
| `config.WithCredentialsProvider(p)` | use credentials provider `p` |

The two file options take a *slice* of paths, even when you have only one.
`[]string{x}` builds a slice holding just `x`.

### Static credentials

For keys you already hold (for example from a secret store), package
`credentials` (`github.com/aws/aws-sdk-go-v2/credentials`) makes a provider:

```go
prov := credentials.NewStaticCredentialsProvider("AKIABILLING0EXAMPLE", "billingSecret", "")
```

The three arguments are key ID, secret, session token, in that order. Pass
`""` as the session token when you don't have one. You pass the provider
itself into an option; you don't call anything on it yourself.

### Options you only sometimes want

The options parameter is *variadic* (as in *Variadic functions*). Each option
has the type `func(*config.LoadOptions) error`. When an option is only
sometimes needed, collect options in a slice and spread it with `...`:

```go
opts := []func(*config.LoadOptions) error{
	config.WithSharedConfigProfile("billing"),
}
if debug {
	opts = append(opts, config.WithClientLogMode(aws.LogRequest))
}
cfg, err := config.LoadDefaultConfig(ctx, opts...)
```

> **Why is it written like this?** Go requires the `...` after `opts` to
> spread a slice into a variadic parameter. The long type
> `func(*config.LoadOptions) error` is required too: it is what each option
> really is. The name `opts` is convention.

`LoadDefaultConfig` returns an error if, for example, the profile you named
does not exist.

### Who wins?

| Priority | Region source |
|---|---|
| 1 (highest) | `config.WithRegion` option in your code |
| 2 | `AWS_REGION` environment variable |
| 3 | `region =` in the profile |

The SDK applies this order for you. You only decide whether to pass
`WithRegion` at all.

### Reading the credentials

`cfg.Credentials.Retrieve(ctx)` returns an `aws.Credentials` struct with
`AccessKeyID`, `SecretAccessKey` and `SessionToken` fields, or an error (for
example when the key ID or secret is empty):

```go
creds, err := cfg.Credentials.Retrieve(ctx)
if err != nil {
	return err
}
log.Println("expires:", creds.Expires) // never log creds.SecretAccessKey
```

Strings can be sliced too: `s[:3]` is the first 3 bytes of `s`. Slicing past
the end panics, so check `len(s)` first.

## Your task

Your deploy tool needs to load AWS settings for the right environment, and
print a startup line that doesn't leak credentials. The tests use the fake
files in `testdata/`, never your real `~/.aws`.

### 1. Build a config from static keys

**Goal:** `StaticConfig(ctx, region, keyID, secret)` returns an `aws.Config`
for `region` that uses `keyID` and `secret` as its credentials.

There is no session token. Return whatever `config.LoadDefaultConfig` returns
(the config and the error). Add the import
`"github.com/aws/aws-sdk-go-v2/credentials"`.

```go
cfg, _ := StaticConfig(ctx, "us-east-1", "AKIAIOSFODNN7EXAMPLE", "wJal...")
cfg.Region // => "us-east-1"
// cfg.Credentials.Retrieve(ctx) gives AccessKeyID "AKIAIOSFODNN7EXAMPLE", SecretAccessKey "wJal..."
```

**Think it through:**
- Which two rows of the options table do you need?
- Which of the three parameters go into the credentials provider, and what goes in the third slot?

### 2. Load a profile from given files

**Goal:** `Load(ctx, s)` returns the config for profile `s.Profile`, read from
the files `s.ConfigFile` and `s.CredentialsFile`; a non-empty `s.Region`
overrides the region.

The starter already puts the profile option in `opts`. Always add the two file
options. Add the region option only when `s.Region` is not `""`. An unknown
profile must return the error from `LoadDefaultConfig`.

With `testdata/team/config` and `testdata/team/credentials`:

| Profile | `s.Region` | `AWS_REGION` | Region you get | Access key ID |
|---|---|---|---|---|
| `staging` | `""` | unset | `eu-west-1` | `AKIASTAGING000EXAMPLE` |
| `prod` | `""` | unset | `us-west-2` | `AKIAPROD00000EXAMPLE` |
| `prod` | `ca-central-1` | unset | `ca-central-1` | `AKIAPROD00000EXAMPLE` |
| `staging` | `""` | `ap-northeast-1` | `ap-northeast-1` | `AKIASTAGING000EXAMPLE` |
| `staging` | `sa-east-1` | `ap-northeast-1` | `sa-east-1` | `AKIASTAGING000EXAMPLE` |
| `qa` or `dev` | | | error | |

**Think it through:**
- Which fields of `s` hold the two file paths, and what type do the file options want?
- Why must the region option be skipped when `s.Region` is `""`? (Look at the *Who wins?* table.)

### 3. Describe the config safely

**Goal:** `Redact(secret)` hides most of a string, and `Describe(ctx, cfg)`
returns a log line that shows the region and the redacted access key ID.

`Redact` returns the first 4 characters followed by exactly `****`. If the
string has 4 characters or fewer (including `""`), it returns just `****`.

```go
Redact("AKIAIOSFODNN7EXAMPLE") // => "AKIA****"
Redact("ASIAXYZ")              // => "ASIA****"
Redact("abcd")                 // => "****"
Redact("")                     // => "****"
```

`Describe` retrieves the credentials from `cfg` and returns exactly
`region=<region> access_key=<redacted key ID>`, with one space between the two
parts. If retrieving fails, return `""` and that error. The secret must never
appear. Add the import `"fmt"` if you format with `fmt.Sprintf`.

```go
Describe(ctx, cfg) // => "region=eu-west-1 access_key=AKIA****", nil
```

**Think it through:**
- In `Redact`, what must you check before slicing, so a short string doesn't panic?
- In `Describe`, which parts are fixed text (`region=`, ` access_key=`) and which come from `cfg` and the credentials?
- Which credential field goes through `Redact`, and which fields don't appear at all?

## Learn more

- [AWS: Configuration and credential file settings](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-files.html)
- [AWS SDK for Go v2: Configure the SDK](https://docs.aws.amazon.com/sdk-for-go/v2/developer-guide/configure-gosdk.html)
- [pkg.go.dev: aws-sdk-go-v2/config](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/config)
- [pkg.go.dev: aws-sdk-go-v2/credentials](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/credentials)
