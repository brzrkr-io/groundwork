# AWS SDK configuration: many regions, no leaked keys

An inventory job has to look at every region your team uses. You don't load
the config ten times: you load it once and make one copy per region.

## Learn

Recap from *AWS SDK configuration: regions, profiles and credentials*:
`config.LoadDefaultConfig(ctx, opts...)`
builds an `aws.Config`. The region comes from, in order, the
`config.WithRegion` option, the `AWS_REGION` environment variable, then the
profile's `region =` line. `cfg.Credentials.Retrieve(ctx)` returns the keys,
and logs may show only a shortened access key ID, never the secret or the
session token.

New: `cfg.Copy()` returns a copy of an `aws.Config`. Changing the copy's
`Region` leaves the original alone, and both copies keep the same credentials
provider, so the keys are not loaded again:

```go
euCfg := cfg.Copy()
euCfg.Region = "eu-west-1"
// cfg.Region is still what it was; euCfg talks to eu-west-1
```

A *session token* is only present for temporary credentials (for example from
an SSO login or an assumed role). Long-lived keys have an empty `SessionToken`.

## Your task

The nightly inventory job loads one profile and then scans several regions.
Two of its functions have bugs that a code review caught.

### 1. One config per region

Complete `ForRegions`. For each entry in `regions`, append `cfg.Copy()` with
`Region` set to that entry. Keep the order of `regions`. An empty `regions`
gives an empty result.

```go
ForRegions(cfg, []string{"us-east-1", "eu-west-1"})
// => two configs: Region "us-east-1", then Region "eu-west-1",
//    both with cfg's credentials
```

### 2. Fix the ignored region override

`LoadProfile(ctx, profile, region)` should let a non-empty `region` win over
everything. Right now it only uses `region` when nothing else set one, so
`--region us-west-1` silently runs against the profile's region. Fix it so
the full precedence holds (an empty `region` changes nothing). An unknown
profile still returns an error.

With the files in `testdata/`:

| Profile | `region` | `AWS_REGION` | Region you get |
|---|---|---|---|
| `ci` | `""` | unset | `us-east-2` |
| `ops` | `""` | unset | `eu-central-1` |
| `ci` | `us-west-1` | unset | `us-west-1` |
| `ops` | `""` | `ap-south-1` | `ap-south-1` |
| `ops` | `eu-north-1` | `ap-south-1` | `eu-north-1` |
| `qa` | any | | error |

### 3. Stop leaking credentials

`Summary` writes the secret key and session token into the deploy log. Fix it
so it returns exactly:

```text
region=<cfg.Region> key=<mask(AccessKeyID)> temporary=<yes|no>
```

`temporary` is `yes` when `SessionToken` is not `""`, otherwise `no`. The key
is already passed through the `mask` helper; keep that.

```go
Summary(ctx, cfg) // => "region=us-east-1 key=AKIA**** temporary=no"
```

## Learn more

- [pkg.go.dev: aws.Config.Copy](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/aws#Config.Copy)
- [AWS: Regions and Availability Zones](https://docs.aws.amazon.com/global-infrastructure/latest/regions/aws-regions.html)
- [AWS SDK for Go v2: Specify the AWS Region](https://docs.aws.amazon.com/sdk-for-go/v2/developer-guide/configure-gosdk.html#specifying-the-aws-region)
