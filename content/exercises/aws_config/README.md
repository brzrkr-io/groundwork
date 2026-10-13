# AWS SDK configuration: regions, profiles and credentials

Every program that talks to AWS starts the same way: work out *which region*
to talk to and *who you are*. Get this right, and never print your secret key
while doing it.

## Learn

### AWS in one minute

*AWS* (Amazon Web Services) rents out servers, storage and databases through an
API. Your resources live in an *account*. Each resource sits in a *region*, a
group of data centres with a name like `us-east-1` or `eu-west-1`. A request
to `eu-west-1` only sees what is in `eu-west-1`.

To call the API you need *credentials*: an *access key ID* (`AKIA...`, safe to
show), a *secret access key* (a password: never log it) and, for temporary
credentials, a *session token*.

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

The SDK module is `github.com/aws/aws-sdk-go-v2`. Every client is built from
an `aws.Config` (package `aws`), which holds the `Region` and a `Credentials`
provider. `config.LoadDefaultConfig` (package `config`) builds one from
your environment and the shared files. You tweak it with *functional
options*: small functions that each change one setting.

```go
cfg, err := config.LoadDefaultConfig(ctx,
	config.WithRegion("eu-west-1"),
	config.WithSharedConfigProfile("staging"),
)
```

| Option | Effect |
|---|---|
| `config.WithRegion(r)` | use region `r` |
| `config.WithSharedConfigProfile(p)` | use profile `p` instead of `default` |
| `config.WithSharedConfigFiles([]string{path})` | read this config file instead of `~/.aws/config` |
| `config.WithSharedCredentialsFiles([]string{path})` | read this credentials file instead of `~/.aws/credentials` |
| `config.WithCredentialsProvider(p)` | use credentials provider `p` |

For keys you already have (for example from a secret store),
`credentials.NewStaticCredentialsProvider(keyID, secret, sessionToken)` from
package `credentials` makes a provider. Pass `""` when there is no session token.

The options parameter is *variadic* (as in *Variadic functions*). Each
option has the type `func(*config.LoadOptions) error`. When some options are
optional, collect them in a slice of that type and spread it with `opts...`:

```go
opts := []func(*config.LoadOptions) error{
	config.WithSharedConfigProfile("staging"),
}
if region != "" {
	opts = append(opts, config.WithRegion(region))
}
cfg, err := config.LoadDefaultConfig(ctx, opts...)
```

`LoadDefaultConfig` returns an error if, for example, the profile you named
does not exist.

### Who wins?

| Priority | Region source |
|---|---|
| 1 (highest) | `config.WithRegion` option in your code |
| 2 | `AWS_REGION` environment variable |
| 3 | `region =` in the profile |

### Reading the credentials

`cfg.Credentials.Retrieve(ctx)` returns an `aws.Credentials` struct with
`AccessKeyID`, `SecretAccessKey` and `SessionToken` fields, or an error (for
example when the key ID or secret is empty):

```go
creds, err := cfg.Credentials.Retrieve(ctx)
if err != nil {
	return "", err
}
fmt.Println(creds.AccessKeyID) // fine; never print creds.SecretAccessKey
```

Even the key ID is better shortened in logs. You can slice a string like a
slice: `s[:4]` is its first 4 bytes.

## Your task

Your deploy tool needs to load AWS settings for the right environment, and
print a startup line that doesn't leak credentials. The tests use the fake
files in `testdata/`, never your real `~/.aws`.

### 1. Build a config from static keys

Complete `StaticConfig` so it calls `config.LoadDefaultConfig` with
`WithRegion(region)` and `WithCredentialsProvider` using
`credentials.NewStaticCredentialsProvider(keyID, secret, "")`. Return what
`LoadDefaultConfig` returns.

```go
cfg, _ := StaticConfig(ctx, "us-east-1", "AKIAIOSFODNN7EXAMPLE", "wJal...")
cfg.Region // => "us-east-1"
```

### 2. Load a profile from given files

Complete `Load`. The starter already builds an options slice with
`WithSharedConfigProfile(s.Profile)` and passes it to `LoadDefaultConfig`.
Add `WithSharedConfigFiles` with `s.ConfigFile` and
`WithSharedCredentialsFiles` with `s.CredentialsFile`, and append
`WithRegion(s.Region)` only when `s.Region` is not `""`. Return the error
for an unknown profile.

With `testdata/team/config` and `testdata/team/credentials`:

| Profile | `s.Region` | `AWS_REGION` | Region you get |
|---|---|---|---|
| `staging` | `""` | unset | `eu-west-1` |
| `prod` | `""` | unset | `us-west-2` |
| `prod` | `ca-central-1` | unset | `ca-central-1` |
| `staging` | `""` | `ap-northeast-1` | `ap-northeast-1` |
| `staging` | `sa-east-1` | `ap-northeast-1` | `sa-east-1` |
| `qa` or `dev` | | | error |

### 3. Describe the config safely

Complete `Redact`: return the first 4 characters followed by exactly
`****`. If the string has 4 characters or fewer, return just `****`.

```go
Redact("AKIAIOSFODNN7EXAMPLE") // => "AKIA****"
Redact("abcd")                 // => "****"
```

Then complete `Describe`: retrieve the credentials from `cfg`, and return
`region=<cfg.Region> access_key=<Redact(AccessKeyID)>`. If `Retrieve` fails,
return `""` and its error. The secret must never appear.

```go
Describe(ctx, cfg) // => "region=eu-west-1 access_key=AKIA****"
```

## Learn more

- [AWS: Configuration and credential file settings](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-files.html)
- [AWS SDK for Go v2: Configure the SDK](https://docs.aws.amazon.com/sdk-for-go/v2/developer-guide/configure-gosdk.html)
- [pkg.go.dev: aws-sdk-go-v2/config](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/config)
- [pkg.go.dev: aws-sdk-go-v2/credentials](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/credentials)
