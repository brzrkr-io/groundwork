## Task 1: Define the settings and build the launch command

First fill the empty `const ()` block with `Region` and `InstanceType`, one per
line. Then build the command by joining fixed text with your constants and the
`imageID` parameter.

### Closer

The constants' values go in quotes inside the block. In the command, fixed
text (including the spaces around each flag) goes inside quotes; constants and
parameters don't:

```go
const (
	Region       = "eu-west-1"
	InstanceType = "t3.small"
)

return "aws ec2 run-instances --region " + Region + " --instance-type " + ...
```

### Answer

```go
const (
	Region       = "eu-west-1"
	InstanceType = "t3.small"
)

return "aws ec2 run-instances --region " + Region + " --instance-type " + InstanceType + " --image-id " + imageID
```

## Task 2: Build the terminate command

This works like the launch command: fixed text joined with `Region` and the
`instanceID` parameter. Don't type `eu-west-1`; use the constant.

### Closer

Fixed text, with a space before and after each flag, goes inside quotes;
`Region` and `instanceID` don't:
`"aws ec2 terminate-instances --region " + Region + ...`.

### Answer

```go
return "aws ec2 terminate-instances --region " + Region + " --instance-ids " + instanceID
```

## Task 3: Link to the console

The URL uses `Region` twice, with fixed text before, between and after. Split
the URL template at each `<...>` to find the pieces.

### Closer

Fixed text goes inside quotes; `Region` and `instanceID` don't. A long
expression may continue on the next line, as long as the line ends with `+`:

```go
return "https://" + Region + ".console.aws.amazon.com/ec2/home?region=" + Region +
	...
```

### Answer

```go
return "https://" + Region + ".console.aws.amazon.com/ec2/home?region=" + Region +
	"#InstanceDetails:instanceId=" + instanceID
```
