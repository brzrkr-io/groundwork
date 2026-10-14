# AWS SDK configuration: many regions, no leaked keys

An inventory job has to look at every region your team uses. You don't load
the config ten times: you load it once and make one copy per region.

**New here:** `aws.Config.Copy`; practice with options, session tokens and safe logging.

## Learn

### Recap

From *AWS SDK configuration: regions, profiles and credentials*:

- `config.LoadDefaultConfig(ctx, opts...)` builds an `aws.Config`.
- The region comes from, in order: the `config.WithRegion` option, the
  `AWS_REGION` environment variable, then the profile's `region =` line.
- `cfg.Credentials.Retrieve(ctx)` returns the keys. Logs may show only a
  shortened access key ID, never the secret or the session token.

### Copying a config

`cfg.Copy()` returns a copy of an `aws.Config`. Changing a field on the copy
leaves the original alone. Both keep the same credentials provider, so keys
are not loaded again:

```go
backup := cfg.Copy()
backup.RetryMaxAttempts = 10
// cfg.RetryMaxAttempts is unchanged
```

> **Why is it written like this?** `Copy` is a method the SDK provides; you
> must call it with `()`. Setting a field with `x.Field = value` is required
> Go syntax. The variable name `backup` is your choice.

### Patching after loading vs. passing an option

You can change `cfg.Region` after loading, but then you must redo the
precedence rules yourself. Passing an option to `LoadDefaultConfig` lets the
SDK apply them. Only add the option when you actually have a value:

```go
opts := []func(*config.LoadOptions) error{config.WithSharedConfigProfile(p)}
if retries > 0 {
	opts = append(opts, config.WithRetryMaxAttempts(retries))
}
```

### Session tokens

A *session token* comes with temporary credentials from an SSO login or an
assumed role, as in
*Assume role: trust policies and temporary credentials*.
Long-lived keys have an empty `SessionToken`. It is as secret as the secret
key.

### Turning a condition into a word

A common pattern: start with a default value, and change it in an `if`:

```go
status := "down"
if healthy {
	status = "up"
}
```

## Your task

The nightly inventory job loads one profile and then scans several regions.
Two of its functions have bugs that a code review caught.

### 1. One config per region

**Goal:** `ForRegions(cfg, regions)` returns one config per entry in
`regions`, each a copy of `cfg` with `Region` set to that entry.

Keep the order of `regions`. Every copy keeps `cfg`'s credentials. An empty
`regions` gives an empty result. `cfg` itself must not change.

```go
ForRegions(cfg, []string{"us-east-1", "eu-west-1"})
// => two configs: Region "us-east-1", then Region "eu-west-1",
//    both with cfg's credentials
```

**Think it through:**
- What do you need to do once per region? (copy, change one field, add to `out`)
- Why make the copy inside the loop, not once before it?

### 2. Fix the ignored region override

**Goal:** `LoadProfile(ctx, profile, region)` lets a non-empty `region` win
over `AWS_REGION` and the profile's region.

Right now it only uses `region` when nothing else set one, so
`--region us-west-1` silently runs against the profile's region. An empty
`region` changes nothing. An unknown profile still returns an error.

With the files in `testdata/`:

| Profile | `region` | `AWS_REGION` | Region you get |
|---|---|---|---|
| `ci` | `""` | unset | `us-east-2` |
| `ops` | `""` | unset | `eu-central-1` |
| `ci` | `us-west-1` | unset | `us-west-1` |
| `ops` | `eu-north-1` | unset | `eu-north-1` |
| `ops` | `""` | `ap-south-1` | `ap-south-1` |
| `ops` | `eu-north-1` | `ap-south-1` | `eu-north-1` |
| `qa` | any | | error |

**Think it through:**
- Look at the `if` in the starter: which value does it check, and which should it check?
- Which option makes the SDK apply the precedence for you, and when should you leave it out?

### 3. Stop leaking credentials

**Goal:** `Summary(ctx, cfg)` returns a log line with no secret key and no
session token in it.

It must return exactly this, with single spaces between the parts:

```text
region=<cfg.Region> key=<mask(AccessKeyID)> temporary=<yes|no>
```

`temporary` is `yes` when `SessionToken` is not `""`, otherwise `no`. Keep
passing the key ID through the `mask` helper.

```go
Summary(ctx, cfg) // long-lived keys  => "region=us-east-1 key=AKIA**** temporary=no"
Summary(ctx, cfg) // temporary keys   => "region=eu-west-1 key=ASIA**** temporary=yes"
```

**Think it through:**
- Which two values in the current `fmt.Sprintf` must go?
- How do you turn "is the token empty?" into the word `yes` or `no` before formatting?

## Learn more

- [pkg.go.dev: aws.Config.Copy](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/aws#Config.Copy)
- [AWS: Regions and Availability Zones](https://docs.aws.amazon.com/global-infrastructure/latest/regions/aws-regions.html)
- [AWS SDK for Go v2: Specify the AWS Region](https://docs.aws.amazon.com/sdk-for-go/v2/developer-guide/configure-gosdk.html#specifying-the-aws-region)
