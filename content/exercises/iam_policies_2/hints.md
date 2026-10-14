## Task 1: Let the backup job upload

One action, one folder of one bucket. Write it as narrowly as the Goal says:
name the action exactly, and end the resource with the folder and a `/*`.

### Closer

One statement in the `Statement` list, with `"Effect": "Allow"`,
`"Action": "s3:PutObject"` and an object ARN: `arn:aws:s3:::` + bucket +
`/nightly/*`. Then delete the whole `"Id": ...` line, including its comma.

### Answer

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "UploadNightlyBackups",
      "Effect": "Allow",
      "Action": "s3:PutObject",
      "Resource": "arn:aws:s3:::acme-db-backups/nightly/*"
    }
  ]
}
```

## Task 2: Let the log reader read prod logs

Listing is something you do to the bucket; reading is something you do to an
object. Look at the table in Learn: each action needs its own kind of ARN.

### Closer

Two statements. The first allows `s3:ListBucket` on the bucket ARN, with no
slash: `arn:aws:s3:::acme-app-logs`. The second allows `s3:GetObject` on the
objects under `prod/`: the bucket ARN plus `/prod/*`.

### Answer

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "ListLogBucket",
      "Effect": "Allow",
      "Action": "s3:ListBucket",
      "Resource": "arn:aws:s3:::acme-app-logs"
    },
    {
      "Sid": "ReadProdLogs",
      "Effect": "Allow",
      "Action": "s3:GetObject",
      "Resource": "arn:aws:s3:::acme-app-logs/prod/*"
    }
  ]
}
```

## Task 3: Let the deploy bot restart servers from the office

Two statements again: the three instance actions on instance ARNs in one
region and account, and `ec2:DescribeInstances` on `"*"`. Both get the same
condition, because the describe call must not work from home either.

### Closer

The resource is `arn:aws:ec2:eu-west-1:111122223333:instance/*`: region and
account are fixed, only the instance ID is a wildcard. The condition is
`"Condition": {"IpAddress": {"aws:SourceIp": "203.0.113.0/24"}}`; put it in
each statement.

### Answer

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "StartStopRebootFromOffice",
      "Effect": "Allow",
      "Action": ["ec2:StartInstances", "ec2:StopInstances", "ec2:RebootInstances"],
      "Resource": "arn:aws:ec2:eu-west-1:111122223333:instance/*",
      "Condition": {"IpAddress": {"aws:SourceIp": "203.0.113.0/24"}}
    },
    {
      "Sid": "DescribeFromOffice",
      "Effect": "Allow",
      "Action": "ec2:DescribeInstances",
      "Resource": "*",
      "Condition": {"IpAddress": {"aws:SourceIp": "203.0.113.0/24"}}
    }
  ]
}
```
