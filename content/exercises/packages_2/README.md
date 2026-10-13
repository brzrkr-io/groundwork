# Packages: two packages, one name

Platform teams often run on more than one cloud, and both cloud SDKs like to
call their packages `region`. This exercise shows how to import two packages
that share a name.

## Learn

### Recap

A package is a folder of `.go` files that share a package clause. Its import
path is the module path from `go.mod` plus the folder, and you call its
exported names as `name.Func()`. Unused or missing imports stop the build.

### Import aliases

Two folders can hold packages with the same name:

| Folder | Import path | Package clause |
|---|---|---|
| `aws/region/` | `gopherlings.local/packages_2/aws/region` | `package region` |
| `gcp/region/` | `gopherlings.local/packages_2/gcp/region` | `package region` |

Import both and Go can't tell which `region` you mean:

```text
region redeclared in this block
```

Give one or both a local name, called an *import alias*. Write it before the
path:

```go
import (
	awsregion "gopherlings.local/packages_2/aws/region"
	gcpregion "gopherlings.local/packages_2/gcp/region"
)

awsregion.Valid("us-east-1")
```

The alias only changes the name inside that one file. The package itself
still says `package region`. Aliases are lower case with no underscores, and
they usually combine the parent folder and the package name.

You will also see aliases where a package name differs from the last part of
its path. For example, `"github.com/aws/aws-sdk-go-v2/aws"` is used as `aws`,
but Kubernetes code often writes
`metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"` because `v1` alone says very
little.

## Your task

The placement service decides which cloud a region name belongs to. The code
lives in three packages:

```text
.
├── go.mod
├── aws/region/region.go         package region (AWS rules)
├── gcp/region/region.go         package region (GCP rules)
└── placement/placement.go       package placement, uses both
```

### 1. Alias the two region imports

`placement/placement.go` imports both `region` packages and fails with
`region redeclared in this block`. Import them as `awsregion` and
`gcpregion`, then update `Provider` to call `awsregion.Valid` and
`gcpregion.Valid`. The rules inside the two `Valid` functions are already
written:

| Package | A valid name has | Examples |
|---|---|---|
| `aws/region` | three or more dash-separated parts, the last a number | `us-east-1`, `us-gov-west-1` |
| `gcp/region` | exactly two parts, the second ending in a digit | `europe-west4`, `us-central1` |

```go
placement.Provider("us-east-1")    // => "aws"
placement.Provider("europe-west4") // => "gcp"
placement.Provider("on-prem")      // => "unknown"
```

### 2. Add Partition to the AWS package

AWS groups its regions into *partitions*. Add an exported function to
`aws/region/region.go`:

```go
func Partition(name string) string
```

| Region name starts with | Partition |
|---|---|
| `us-gov-` | `aws-us-gov` |
| `cn-` | `aws-cn` |
| anything else | `aws` |

```go
region.Partition("us-gov-west-1") // => "aws-us-gov"
region.Partition("cn-north-1")    // => "aws-cn"
region.Partition("eu-west-2")     // => "aws"
```

`TestPartition` lives in `aws/region` and fails with `undefined: Partition`
until the function exists. You'll need the `strings` package.

### 3. Finish Describe

`Describe` in `placement/placement.go` returns a sentence for the on-call
dashboard. Use `Provider`, and for AWS regions call `awsregion.Partition`:

```go
placement.Describe("us-gov-west-1") // => "us-gov-west-1 is AWS, partition aws-us-gov"
placement.Describe("eu-west-2")     // => "eu-west-2 is AWS, partition aws"
placement.Describe("europe-west4")  // => "europe-west4 is GCP"
placement.Describe("on-prem")       // => "on-prem is not a known region"
```

## Learn more

- [Go spec: Import declarations](https://go.dev/ref/spec#Import_declarations)
- [Effective Go: Package names](https://go.dev/doc/effective_go#package-names)
- [AWS: Partitions](https://docs.aws.amazon.com/whitepapers/latest/aws-fault-isolation-boundaries/partitions.html)
