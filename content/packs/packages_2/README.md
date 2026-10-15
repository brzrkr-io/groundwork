# Packages: two packages, one name

Platform teams often run on more than one cloud, and both cloud SDKs like to
call their packages `region`. This exercise shows how to import two packages
that share a name.

**New here:** import aliases.

## Learn

### Recap

- A package is a folder of `.go` files with the same package clause.
- Its import path is the module path from `go.mod`, plus the folder.
- You call its exported names as `name.Func()`.
- Unused or missing imports stop the build.

### Two packages, one name

Package names only have to be unique per *folder*, not per module. So two
folders can hold packages with the same name. For a module called
`example.com/infra`:

| Folder | Import path | Package clause |
|---|---|---|
| `dns/record/` | `example.com/infra/dns/record` | `package record` |
| `audit/record/` | `example.com/infra/audit/record` | `package record` |

Import both into one file and Go can't tell which `record` you mean:

```text
record redeclared in this block
```

### Import aliases

Give an import its own local name, an *import alias*, by writing the name
before the path:

```go
import (
	dnsrecord "example.com/infra/dns/record"
	auditrecord "example.com/infra/audit/record"
)

func Save(name string) {
	dnsrecord.Add(name)
	auditrecord.Log("added " + name)
}
```

> **Why is it written like this?**
> - **Go requires** the alias before the path, with a space and no `=`.
>   After that, the file must use the alias. The original name `record`
>   stops working in that file.
> - **Go requires** nothing else to change. The alias exists only in this
>   file; the package itself still says `package record`.
> - **Convention:** aliases are lower case, with no underscores. They usually
>   join the parent folder and the package name, like `dnsrecord`.
> - **Convention:** only alias when you must. Here you must, because of the
>   clash.

You'll also see aliases when a package's last path part says very little.
Kubernetes code writes `metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"`,
because `v1` alone could be anything.

### Your own package's helpers

Inside a package, you call its own functions without a prefix. `Describe`
in `placement` calls `Provider`, not `placement.Provider`.

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

**Goal:** `placement/placement.go` imports the AWS package as `awsregion` and the GCP one as `gcpregion`, and `Provider` uses each.

Right now it fails with `region redeclared in this block`. The rules inside
the two `Valid` functions are already written:

| Package | A valid name has | Examples |
|---|---|---|
| `aws/region` | three or more dash-separated parts, the last a number | `us-east-1`, `us-gov-west-1` |
| `gcp/region` | exactly two parts, the second ending in a digit | `europe-west4`, `us-central1` |

`Provider` checks AWS first, then GCP:

```go
placement.Provider("us-east-1")    // => "aws"
placement.Provider("europe-west4") // => "gcp"
placement.Provider("on-prem")      // => "unknown"
```

**Think it through:**
- Which import path ends in `aws/region`, and which in `gcp/region`?
- After aliasing, `region.` no longer works in this file. Which two lines in `Provider` use it?

### 2. Add Partition to the AWS package

**Goal:** add an exported function `Partition` to `aws/region/region.go` that returns the AWS partition of a region name.

AWS groups its regions into *partitions*. The function takes a region name
(`string`) and returns a `string`:

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
until the function exists.

**Think it through:**
- Why must the name start with a capital `P`? Who calls it in task 3?
- Which function in the `strings` package checks how a string *starts*? `region.go` already imports `strings`.
- Inside its own package, the call is just `Partition(...)`, with no prefix.

### 3. Finish Describe

**Goal:** `Describe` in `placement/placement.go` returns one sentence about a region for the on-call dashboard.

Use `Provider` to decide which sentence. For AWS regions, add the partition
from your new `Partition`. The words `AWS` and `GCP` are upper case in the
sentence:

```go
placement.Describe("us-gov-west-1") // => "us-gov-west-1 is AWS, partition aws-us-gov"
placement.Describe("eu-west-2")     // => "eu-west-2 is AWS, partition aws"
placement.Describe("europe-west4")  // => "europe-west4 is GCP"
placement.Describe("on-prem")       // => "on-prem is not a known region"
```

**Think it through:**
- Which part of each sentence is the `name` parameter, and which is fixed text in quotes? Watch the spaces.
- From `placement`, `Partition` lives in another package. Which name do you put in front of it?

## Learn more

- [Go spec: Import declarations](https://go.dev/ref/spec#Import_declarations)
- [Effective Go: Package names](https://go.dev/doc/effective_go#package-names)
- [AWS: Partitions](https://docs.aws.amazon.com/whitepapers/latest/aws-fault-isolation-boundaries/partitions.html)
