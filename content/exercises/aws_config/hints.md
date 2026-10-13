## Task 1: Build a config from static keys

`config.LoadDefaultConfig` takes a context and any number of option functions.
You need two options here: one for the region and one for the credentials.
It already returns `(aws.Config, error)`, so you can return its result directly.

### Closer

Add the import `"github.com/aws/aws-sdk-go-v2/credentials"` (`config` is
already imported). The credentials option wraps a provider:
`config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(keyID, secret, ""))`.

### Answer

```go
return config.LoadDefaultConfig(ctx,
	config.WithRegion(region),
	config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(keyID, secret, "")),
)
```

The empty third argument is the session token, which long-lived keys don't have.

## Task 2: Load a profile from given files

The options slice already holds the profile option. Add the two file options
that always apply, then add the region option only when `s.Region` is set.

### Closer

The file options take a slice of paths:
`config.WithSharedConfigFiles([]string{s.ConfigFile})` and
`config.WithSharedCredentialsFiles([]string{s.CredentialsFile})`. Use
`opts = append(opts, config.WithRegion(s.Region))` inside an `if`.

### Answer

```go
opts := []func(*config.LoadOptions) error{
	config.WithSharedConfigProfile(s.Profile),
	config.WithSharedConfigFiles([]string{s.ConfigFile}),
	config.WithSharedCredentialsFiles([]string{s.CredentialsFile}),
}
if s.Region != "" {
	opts = append(opts, config.WithRegion(s.Region))
}
return config.LoadDefaultConfig(ctx, opts...)
```

When `s.Region` is empty, the SDK already applies `AWS_REGION` and then the profile's region in the right order.

## Task 3: Describe the config safely

`Redact` needs the length of the string and a slice of its first 4
characters. `Describe` asks the config's credentials provider for the keys,
then formats one line with the region and the redacted key ID.

### Closer

`secret[:4]` is the first 4 bytes. Credentials come from
`cfg.Credentials.Retrieve(ctx)`, which returns `(aws.Credentials, error)`.
Add `"fmt"` to the imports for `fmt.Sprintf("region=%s access_key=%s", ...)`.

### Answer

```go
func Redact(secret string) string {
	if len(secret) <= 4 {
		return "****"
	}
	return secret[:4] + "****"
}

func Describe(ctx context.Context, cfg aws.Config) (string, error) {
	creds, err := cfg.Credentials.Retrieve(ctx)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("region=%s access_key=%s", cfg.Region, Redact(creds.AccessKeyID)), nil
}
```

Only `AccessKeyID` goes through `Redact`; the secret is never formatted at all.
