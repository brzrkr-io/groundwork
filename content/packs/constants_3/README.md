# Constants: write your own config block

Your team launches EC2 instances with the AWS CLI, always in the same region
and with the same instance type. Instead of typing those values into every
command, you'll define them once as constants.

**New here:** nothing; practice writing a `const` block and using it.

## Learn

### Recap

- `const ( ... )` groups related constants, one `Name = value` per line.
- Inside the block, lines don't start with `const`; the block already says it.
- You use a constant by its name, outside quotes, like a parameter.

When a value appears in several places, a constant means you change it in
one place and every use follows.

### Filling in an empty block

The starter has an empty block: `const ()`. To fill it, put the `)` on its
own line and add your constants between the brackets:

```go
// Backup settings.
const (
	// Bucket is the S3 bucket backups go to.
	Bucket = "acme-backups"
	// Prefix starts the name of every backup file.
	Prefix = "nightly/"
)
```

> **Why is it written like this?**
> - **Go requires** `Name = value` inside the block, with the value in quotes
>   if it's text.
> - **Convention:** a comment above each constant, starting with its name.
> - **Your choice:** the order of the constants.

### Names versus values in a function

A function's `func` line declares its parameters, like `imageID string`. In
the body, you use just the name:

| In the `func` line | In the body |
|---|---|
| `func Backup(file string) string {` | `file` |
| the function's name, `Backup` | not used to build the result |

Writing `Backup(file string)` or `file string` inside the body won't compile:
those parts belong only in the `func` line.

## Your task

Define the AWS settings, then build the commands and URL your runbook needs.
Every function must use your constants, not the literal text `eu-west-1` or
`t3.small`.

### 1. Define the settings and build the launch command

**Goal:** add `Region` and `InstanceType` to the `const` block, then make
`RunInstancesCommand` return the launch command.

| Name | Value |
|---|---|
| `Region` | `eu-west-1` (text) |
| `InstanceType` | `t3.small` (text) |

The command, with each `<...>` replaced by that constant or parameter:

```text
aws ec2 run-instances --region <Region> --instance-type <InstanceType> --image-id <imageID>
```

```go
RunInstancesCommand("ami-0abc1234")
// => "aws ec2 run-instances --region eu-west-1 --instance-type t3.small --image-id ami-0abc1234"

RunInstancesCommand("ami-09f8e7d6")
// => "aws ec2 run-instances --region eu-west-1 --instance-type t3.small --image-id ami-09f8e7d6"
```

The block is on line 6; the function is below it. Everything outside `<...>`
is fixed text, including the spaces.

**Think it through:**
- The block first: which values need quotes?
- In the command, which three pieces are names, and which are fixed text?
- Does each fixed-text piece start or end with a space?

### 2. Build the terminate command

**Goal:** make `TerminateCommand` return the terminate command.

```text
aws ec2 terminate-instances --region <Region> --instance-ids <instanceID>
```

```go
TerminateCommand("i-0abc1234")
// => "aws ec2 terminate-instances --region eu-west-1 --instance-ids i-0abc1234"

TerminateCommand("i-0fedcba9")
// => "aws ec2 terminate-instances --region eu-west-1 --instance-ids i-0fedcba9"
```

**Think it through:**
- Which parameter does this function have? Use that name.

### 3. Link to the console

**Goal:** make `ConsoleURL` return the console URL, which contains the region
twice.

```text
https://<Region>.console.aws.amazon.com/ec2/home?region=<Region>#InstanceDetails:instanceId=<instanceID>
```

```go
ConsoleURL("i-0abc1234")
// => "https://eu-west-1.console.aws.amazon.com/ec2/home?region=eu-west-1#InstanceDetails:instanceId=i-0abc1234"

ConsoleURL("i-0fedcba9")
// => "https://eu-west-1.console.aws.amazon.com/ec2/home?region=eu-west-1#InstanceDetails:instanceId=i-0fedcba9"
```

There are no spaces in a URL.

**Think it through:**
- Split the URL at each `<...>`. How many fixed-text pieces are left?

## Learn more

- [AWS CLI: run-instances](https://docs.aws.amazon.com/cli/latest/reference/ec2/run-instances.html)
- [AWS CLI: terminate-instances](https://docs.aws.amazon.com/cli/latest/reference/ec2/terminate-instances.html)
- [Go by Example: Constants](https://gobyexample.com/constants)
