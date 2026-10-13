## Task 1: two options

Add the import `"github.com/aws/aws-sdk-go-v2/credentials"` (`config` is
already imported). `LoadDefaultConfig` already
returns `(aws.Config, error)`, so you can return its result directly:

```go
return config.LoadDefaultConfig(ctx,
	config.WithRegion(region),
	config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(keyID, secret, "")),
)
```

## Task 2: a slice of options

Start the slice with the three options that always apply, then append the
region only when it is set:

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

Don't set the region yourself when `s.Region` is empty: the SDK already
applies `AWS_REGION` and the profile's region in the right order.

## Task 3: Redact with a slice expression

```go
if len(secret) <= 4 {
	return "****"
}
return secret[:4] + "****"
```

## Task 3: Describe

Add the import `"fmt"`.

```go
creds, err := cfg.Credentials.Retrieve(ctx)
if err != nil {
	return "", err
}
return fmt.Sprintf("region=%s access_key=%s", cfg.Region, Redact(creds.AccessKeyID)), nil
```
