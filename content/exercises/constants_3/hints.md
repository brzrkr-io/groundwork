## Task 1: fill the block

Inside the parentheses, one constant per line:

```go
const (
	Region       = "eu-west-1"
	InstanceType = "t3.small"
)
```

## Watch the spaces

Each flag needs a space before it and a space after it. For example:
`"aws ec2 run-instances --region " + Region + " --instance-type " + ...`

## Task 3: Region twice

```go
return "https://" + Region + ".console.aws.amazon.com/ec2/home?region=" + Region +
	"#InstanceDetails:instanceId=" + instanceID
```

A long expression may continue on the next line, as long as the line ends
with `+`.
