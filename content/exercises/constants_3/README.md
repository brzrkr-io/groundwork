# Constants: write your own config block

Your team launches EC2 instances with the AWS CLI, always in the same region
and with the same instance type. Instead of typing those values into every
command, you'll define them once as constants.

## Learn

### Recap

`const ( ... )` groups related constants, and a constant is used exactly like
a variable. When the same value appears in several places, a constant means
you change it in one place and every use follows.

This time the constant block is empty. You fill it in, then use it in three
functions.

## Your task

Define the AWS settings, then build the commands and URL your runbook needs.
Every function must use your constants, not the literal text `eu-west-1` or
`t3.small`.

### 1. Define the settings and build the launch command

Add two constants to the `const ()` block:

| Name | Value |
|---|---|
| `Region` | `"eu-west-1"` |
| `InstanceType` | `"t3.small"` |

Then make `RunInstancesCommand` return:

```text
aws ec2 run-instances --region <Region> --instance-type <InstanceType> --image-id <imageID>
```

```go
RunInstancesCommand("ami-0abc1234")
// => "aws ec2 run-instances --region eu-west-1 --instance-type t3.small --image-id ami-0abc1234"

RunInstancesCommand("ami-09f8e7d6")
// => "aws ec2 run-instances --region eu-west-1 --instance-type t3.small --image-id ami-09f8e7d6"
```

### 2. Build the terminate command

```text
aws ec2 terminate-instances --region <Region> --instance-ids <instanceID>
```

```go
TerminateCommand("i-0abc1234")
// => "aws ec2 terminate-instances --region eu-west-1 --instance-ids i-0abc1234"

TerminateCommand("i-0fedcba9")
// => "aws ec2 terminate-instances --region eu-west-1 --instance-ids i-0fedcba9"
```

### 3. Link to the console

The AWS console URL for an instance contains the region twice:

```text
https://<Region>.console.aws.amazon.com/ec2/home?region=<Region>#InstanceDetails:instanceId=<instanceID>
```

```go
ConsoleURL("i-0abc1234")
// => "https://eu-west-1.console.aws.amazon.com/ec2/home?region=eu-west-1#InstanceDetails:instanceId=i-0abc1234"
```

## Learn more

- [AWS CLI: run-instances](https://docs.aws.amazon.com/cli/latest/reference/ec2/run-instances.html)
- [AWS CLI: terminate-instances](https://docs.aws.amazon.com/cli/latest/reference/ec2/terminate-instances.html)
- [Go by Example: Constants](https://gobyexample.com/constants)
