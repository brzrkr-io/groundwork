# ARNs: accounts, regions and resource scope

Every AWS resource has a full name that says who owns it and where it lives:
its *ARN*. IAM policies, CloudTrail logs and audit tools are full of them, so
reading one at a glance is a basic cloud skill. Here you take ARNs and zone
names apart in Go.

**New here:** ARNs, AWS accounts, regions and availability zones, `strings.SplitN`.

## Learn

See *The cloud from the ground up* for the bigger story; this section gives
you what the tasks need.

### Accounts

An *AWS account* is a container for resources, with its own bill and its own
users and permissions. Nothing in one account can touch another account
unless someone allows it on purpose. That makes the account the main
**billing and security boundary**.

Each account has a 12-digit *account ID*, such as `111122223333`. Companies
usually run many accounts: one for production, one for staging, one per team,
so a mistake in one can't break the others.

### Regions and availability zones

A *region* is a part of the world where AWS runs data centres. Its code is
`<area>-<direction>-<number>`:

| Region code | Where |
|---|---|
| `us-east-1` | N. Virginia, USA |
| `eu-west-1` | Ireland |
| `eu-central-2` | Zurich |
| `ap-southeast-2` | Sydney |

Each region has several *availability zones* (AZs): one or more data centres
with their own power and network, kilometres away from the other zones. A
zone's name is the region code plus one letter: `eu-west-1a`, `eu-west-1b`,
`eu-west-1c`. If one zone floods, the others keep running, so you spread
servers across zones.

The letters are shuffled per account: your `us-east-1a` may be a different
building from someone else's `us-east-1a`.

### Global and regional services

Most services are *regional*: an EC2 server (a virtual machine) lives in one
region, and `aws ec2 describe-instances` only shows the region you ask about.
A few are *global*:

| Service | Scope | What that means |
|---|---|---|
| EC2 (servers), Lambda, SQS | regional | each region has its own, separate set |
| IAM (users, roles, policies) | account | one set per account, used in every region |
| S3 bucket names | global | unique across **all** AWS accounts in the world |

An S3 bucket's data does sit in one region, but its name is global: if someone
anywhere has `acme-app-logs`, you can't have it.

### The ARN format

An *Amazon Resource Name* (ARN) has six parts separated by colons:

```text
arn:partition:service:region:account-id:resource
arn:aws      :ec2    :eu-west-1:111122223333:instance/i-0abc1234def567890
```

| Part | Meaning |
|---|---|
| `arn` | always the word `arn` (lowercase) |
| partition | `aws`; `aws-cn` in China, `aws-us-gov` in GovCloud |
| service | `ec2`, `s3`, `iam`, `lambda`, ... |
| region | where it lives; **empty** for IAM and S3 |
| account-id | who owns it; **empty** for S3 (bucket names are global already) |
| resource | the service's own name for it |

The resource part looks different per service: `type/id`, `type:id`, or just
a name. It can contain colons and slashes itself:

| Resource | ARN |
|---|---|
| EC2 instance | `arn:aws:ec2:eu-west-1:111122223333:instance/i-0abc1234def567890` |
| Lambda function | `arn:aws:lambda:us-east-1:123456789012:function:resize-images` |
| CloudWatch log group | `arn:aws:logs:us-east-1:123456789012:log-group:/app/web:*` |
| SQS queue | `arn:aws:sqs:eu-west-1:111122223333:orders` |
| IAM role | `arn:aws:iam::111122223333:role/ci-deployer` |
| S3 bucket | `arn:aws:s3:::acme-app-logs` |
| S3 object | `arn:aws:s3:::acme-app-logs/prod/web.log` |

Empty parts still keep their colons: that's why S3 ARNs have `:::`.

### Splitting with a limit: strings.SplitN

`strings.Split` cuts at **every** separator. That breaks the log group ARN
above into eight pieces. `strings.SplitN(s, sep, n)` stops after `n` pieces,
and the last piece keeps the rest of the string, separators and all:

```go
line := "2026-10-13T09:15:02Z ERROR disk full: /var/log: 100%"
parts := strings.SplitN(line, " ", 3)
// parts[0] = "2026-10-13T09:15:02Z"
// parts[1] = "ERROR"
// parts[2] = "disk full: /var/log: 100%"

strings.SplitN("only-one-piece", " ", 3) // => ["only-one-piece"]
```

If there are fewer separators than `n-1`, you get fewer pieces, so check
`len(parts)` before you index.

> **Why is it written like this?** `n` is the **maximum number of pieces**,
> not the number of cuts: Go defines it that way (`n` of `-1` means no limit,
> like `Split`). Checking `len` before indexing is required in practice:
> `parts[5]` on a shorter slice panics.

### Try it (optional)

Needs an AWS account and credentials; never needed for the tests.

```sh
aws sts get-caller-identity                     # your account ID and your own ARN
aws ec2 describe-regions --output table         # the regions your account can use
aws ec2 describe-availability-zones --region eu-west-1 --output table   # zone names and IDs
aws iam list-roles --query 'Roles[].Arn'        # ARNs of the IAM roles in your account
```

## Your task

You're writing helpers for a cloud audit tool. It reads ARNs from IAM
policies and inventory reports and needs to know where each one lives.

### 1. Parse an ARN

**Goal:** `Parse(s)` splits `s` into an `ARN` struct with `Partition`, `Service`, `Region`, `Account` and `Resource`.

- Split into **at most six** parts at `:`. Everything after the fifth colon is
  the `Resource`, colons included.
- Return `ARN{}` and an error when:
  - there are fewer than six parts;
  - the first part isn't exactly `arn`;
  - the partition, service or resource is empty.
- Region and account may be empty.
- Any error message works; the tests only check that there is one. Add
  `"fmt"` and `"strings"` to the imports.

```go
Parse("arn:aws:logs:us-east-1:123456789012:log-group:/app/web:*")
// => ARN{Partition: "aws", Service: "logs", Region: "us-east-1",
//        Account: "123456789012", Resource: "log-group:/app/web:*"}, nil
Parse("arn:aws:s3:::acme-app-logs")
// => ARN{Partition: "aws", Service: "s3", Region: "", Account: "",
//        Resource: "acme-app-logs"}, nil
Parse("arn:aws:iam:111122223333:role/ci")   // => ARN{}, error (five parts)
Parse("ARN:aws:s3:::acme-app-logs")         // => ARN{}, error (not "arn")
Parse("arn:aws:s3:::")                      // => ARN{}, error (no resource)
```

**Think it through:**
- What should `n` be so the resource keeps its colons?
- Which checks need the parts, and so must come after the length check?

### 2. Tell the scope

**Goal:** `Scope(a)` returns `"regional"`, `"account"` or `"global"`.

| `a.Region` | `a.Account` | Scope | Example |
|---|---|---|---|
| set | set or empty | `"regional"` | EC2 instance; a public machine image, which has no account |
| empty | set | `"account"` | IAM role or user |
| empty | empty | `"global"` | S3 bucket or object |

```go
Scope(ARN{Service: "ec2", Region: "eu-west-1", Account: "111122223333"}) // => "regional"
Scope(ARN{Service: "iam", Account: "111122223333"})                     // => "account"
Scope(ARN{Service: "s3", Resource: "acme-app-logs"})                    // => "global"
```

**Think it through:**
- Which field decides `"regional"` on its own? Check it first.

### 3. Find a zone's region

**Goal:** `RegionOfZone(zone)` returns the region a zone name belongs to: `"us-east-1a"` gives `"us-east-1"`.

A zone name is a region code plus one letter. Return `""` and an error unless:

- the **last character** is a lowercase letter `a`–`z`; and
- what's left (the region) contains **at least two** `-` and **ends in a
  digit** `0`–`9`.

```go
RegionOfZone("us-east-1a")      // => "us-east-1", nil
RegionOfZone("ap-southeast-2b") // => "ap-southeast-2", nil
RegionOfZone("us-gov-west-1a")  // => "us-gov-west-1", nil
RegionOfZone("eu-west-1")       // => "", error (a region, not a zone)
RegionOfZone("us-east-1A")      // => "", error (capital letter)
RegionOfZone("us-east1a")       // => "", error (only one dash)
RegionOfZone("eu-west-a")       // => "", error ("eu-west-" doesn't end in a digit)
RegionOfZone("")                // => "", error
```

**Think it through:**
- `zone[len(zone)-1]` is the last byte. What happens to that expression when
  `zone` is empty?
- Which function from *The strings package* counts the dashes?

### 4. Find resources in other accounts

**Goal:** `ForeignResources(account, arns)` returns the ARNs owned by an account other than `account`.

This is the first step of a cross-account access review: "which resources in
this policy belong to someone else?"

- Parse every ARN with your `Parse`. If one fails, return `nil` and its error.
- Keep an ARN when its account is set **and** differs from `account`. ARNs
  without an account (S3) are never foreign.
- Return the original strings, in input order. Nothing foreign gives an empty
  result (`nil` is fine).

```go
ForeignResources("111122223333", []string{
	"arn:aws:ec2:eu-west-1:111122223333:instance/i-0abc1234def567890",
	"arn:aws:s3:::acme-app-logs",
	"arn:aws:iam::444455556666:role/auditor",
	"arn:aws:sqs:eu-west-1:123456789012:orders",
})
// => ["arn:aws:iam::444455556666:role/auditor" "arn:aws:sqs:eu-west-1:123456789012:orders"], nil
```

**Think it through:**
- Which two conditions must both hold to keep an ARN?
- Do you append the parsed `ARN` or the string you were given?

## Learn more

- [Identify AWS resources with Amazon Resource Names (ARNs)](https://docs.aws.amazon.com/IAM/latest/UserGuide/reference-arns.html)
- [AWS Regions](https://docs.aws.amazon.com/global-infrastructure/latest/regions/aws-regions.html)
- [EC2: Regions and Zones](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-regions-availability-zones.html)
- [strings.SplitN](https://pkg.go.dev/strings#SplitN)
