# The cloud from the ground up

**In one sentence:** a public cloud like AWS rents you servers, disks, networks and managed services through an API, billed by the hour or the gigabyte, inside accounts and regions you choose.

**Why it matters:** a platform engineer's day is full of cloud questions: "why can't this instance reach the database?", "which account is that bucket in?", "who deleted the load balancer?", "why did the bill double?". Every one has a short answer once you know where things live, how they connect and how they're paid for.

**You'll learn:**

- What "the cloud" really is, and which parts of security are yours and which are AWS's.
- Where resources live: accounts, regions and availability zones, and how an ARN names them.
- The handful of building blocks most systems use: EC2, EBS, S3, RDS, load balancers, IAM.
- How a VPC network is put together: subnets, route tables, gateways, security groups and network ACLs.
- How you talk to AWS (console, CLI, SDK, infrastructure as code) and how CloudTrail records it.
- How the bill is calculated, and how to see a surprise coming.

## Someone else's data centre, behind an API

Before the cloud, getting a server meant ordering hardware, waiting weeks, racking it and wiring it. In a cloud, you make an API call and a virtual machine is running a minute later. When you're done, another call deletes it and the billing stops.

That's the whole idea: **infrastructure you rent by API call, and pay for while you use it.** Underneath it's still real machines in real buildings; AWS runs them, you rent slices.

Clouds sell at different levels:

| Level | You get | You manage | AWS example |
|---|---|---|---|
| Infrastructure (IaaS) | Virtual machines, disks, networks | OS, patches, your app | EC2, EBS, VPC |
| Platform (PaaS) | A place to run code or data | Your app and its config | RDS, EKS, Lambda |
| Software (SaaS) | A finished product | Your users and data | WorkMail, QuickSight |

### Who secures what

AWS calls the split the *shared responsibility model*: AWS is responsible for security **of** the cloud (buildings, hardware, the hypervisor, its own services); you're responsible for security **in** the cloud.

| AWS's job | Your job |
|---|---|
| Guards, power, disks being wiped | Who can log in to your account |
| Patching the hosts under your VMs | Patching the OS inside your EC2 instances |
| S3 storing bytes durably | Not making your bucket public by mistake |
| The network between data centres | Which ports your security groups open |

Almost every cloud breach in the news is on the right-hand side: a leaked key, a public bucket, an open port. That's good news: it's the side you control.

## Where things live: accounts, regions and zones

### Accounts

An *AWS account* is a container for resources with its own bill and its own users and permissions. It's identified by a 12-digit number such as `111122223333`.

The account is the strongest wall AWS has. Nothing in one account can touch another unless someone explicitly allows it. So companies use **many accounts**: one per environment (dev, staging, prod) or per team, grouped under *AWS Organizations*, which also puts all the bills together.

```text
Organization (acme)
├── 111122223333  prod
├── 444455556666  staging
└── 123456789012  security (audit logs land here)
```

A bug in a dev script can't delete production if production is a different account.

### Regions and availability zones

A *region* is a geographic area with its own complete copy of AWS: `us-east-1` (North Virginia), `eu-west-1` (Ireland), `ap-southeast-2` (Sydney). There are more than 30. Regions are independent: a resource in `eu-west-1` doesn't appear when you list `us-east-1`, and data stays in the region you put it in unless you copy it out.

Each region contains several *availability zones* (AZs): one or more data centres with their own power, cooling and network, a few kilometres to tens of kilometres from the others. An AZ's name is the region plus a letter: `eu-west-1a`, `eu-west-1b`, `eu-west-1c`.

```text
region eu-west-1
├── eu-west-1a   (data centres)   ─┐
├── eu-west-1b   (data centres)   ─┼─ fast private links between them
└── eu-west-1c   (data centres)   ─┘
```

Why it matters: **one AZ can fail** (a power cut, a flood, a bad network change) without taking the others down. So you run at least two copies of anything important, in different AZs. A whole region failing is much rarer.

One surprise: the letters are shuffled per account. Your `us-east-1a` may be a different building from another account's `us-east-1a`. AWS also gives each zone a stable *AZ ID* like `use1-az1` for when two accounts must agree.

### Scope: global, regional and zonal

Every resource has a scope, and it decides where you look for it:

| Scope | Examples | What it means |
|---|---|---|
| Global | IAM users and roles, S3 bucket *names*, CloudFront | One per account (or one in the whole world, for bucket names) |
| Regional | VPCs, S3 buckets' data, load balancers, RDS clusters | Lives in one region |
| Zonal | EC2 instances, EBS volumes, subnets | Lives in one AZ |

An EBS disk in `eu-west-1a` can only attach to an instance in `eu-west-1a`. That one rule explains many "why can't I attach this volume?" questions.

### ARNs: one name for everything

An *Amazon Resource Name* (ARN) names any resource uniquely. It's six parts separated by colons:

```text
arn:aws:ec2:eu-west-1:111122223333:instance/i-0abcd1234ef567890
 │   │   │      │          │        └─ resource (type/id)
 │   │   │      │          └─ account
 │   │   │      └─ region
 │   │   └─ service
 │   └─ partition (aws, or aws-cn / aws-us-gov)
 └─ always "arn"
```

Global things leave parts empty: an IAM role has no region (`arn:aws:iam::111122223333:role/deployer`), and an S3 bucket has neither region nor account (`arn:aws:s3:::acme-logs`). You'll parse ARNs in Go in *ARNs: accounts, regions and resource scope*.

## The building blocks

AWS has hundreds of services. Most systems are built from a few:

| Service | What it is | The nearest thing you know |
|---|---|---|
| EC2 | Virtual machines ("instances") | A Linux server |
| EBS | Network disks for EC2, in one AZ | A disk you can unplug and re-attach |
| S3 | Object storage: files ("objects") in buckets, addressed by key | A huge key-value store for files |
| RDS | Managed databases (PostgreSQL, MySQL) | The database from *Databases and SQL*, patched and backed up for you |
| ELB | Load balancers | The balancer from *Load balancing: round robin with health checks* |
| EKS | Managed Kubernetes control plane | The cluster from *Kubernetes fundamentals* |
| Lambda | Run a function per event, no server to manage | A handler that only exists while called |
| IAM | Identities and permissions | Users, groups and file modes, for the whole cloud |
| CloudWatch | Metrics, logs and alarms | The metrics from *Production services* |

**Managed** means AWS runs the boring parts (patching, backups, failover) and you pay more per hour for it. For a small team that trade is usually worth it.

## Networks: VPCs, subnets and the doors in and out

A *VPC* (virtual private cloud) is your own private network inside one region, with an address range you pick, like `10.0.0.0/16`, written in the CIDR notation from *IP addresses: classify, allowlist, overlap*. Every region in a new account comes with a *default VPC* (`172.31.0.0/16`) so you can start quickly; real systems usually build their own.

A *subnet* is a slice of the VPC's range that lives in **one AZ**. You put each instance in a subnet, and so in an AZ.

```text
VPC 10.0.0.0/16 (eu-west-1)
├── eu-west-1a
│   ├── public-a    10.0.0.0/20    load balancer
│   └── private-a   10.0.32.0/20   app servers, database
└── eu-west-1b
    ├── public-b    10.0.16.0/20   load balancer
    └── private-b   10.0.48.0/20   app servers, database replica
```

Each subnet loses five addresses to AWS (the network address, the VPC router, DNS, one reserved for later, and the last one), so a `/24` gives you 251 usable addresses, not 256. You'll plan subnets like these in *VPC planning: subnets per availability zone*.

### Public and private subnets

What makes a subnet public or private is just its *route table*, the list that tells packets where to go next:

| Subnet | Route for `0.0.0.0/0` (everything else) | Result |
|---|---|---|
| Public | An *internet gateway* | Instances with a public IP can be reached from the internet and reach out |
| Private | A *NAT gateway* in a public subnet | Instances can reach out (for updates, APIs) but nothing can connect in |
| Isolated | No route | Only traffic inside the VPC |

The usual design: load balancers in public subnets, everything else in private ones.

### Security groups and network ACLs

Two firewalls guard traffic, and they behave differently:

| | Security group | Network ACL |
|---|---|---|
| Attached to | An instance's network interface | A subnet |
| Rules | Allow only | Allow and deny, numbered, first match wins |
| Replies | Allowed automatically (*stateful*) | Must be allowed explicitly (*stateless*) |
| Default | Deny all inbound, allow all outbound | The default ACL allows everything |

"Stateless" catches people out. A client connects from a random high port (say `50432`) to your port 443. The reply goes back to port `50432`. A security group remembers the connection and lets the reply out. A network ACL doesn't remember anything, so its outbound rules must allow ports 1024–65535 back to the client, or every reply is dropped. You'll build a checker for this in *Security groups and network ACLs: will this packet get through?*.

## Talking to AWS

Everything in AWS is an API call: clicking in the web *console*, running the `aws` CLI, calling the SDK from Go, or applying Terraform. Each call is signed with credentials, checked by IAM, and recorded.

The CLI is the quickest way to look around. Read-only commands are safe to try in any account you're allowed into:

```bash
aws sts get-caller-identity                     # who am I, in which account?
aws ec2 describe-availability-zones --region eu-west-1
aws ec2 describe-instances --filters Name=tag:env,Values=prod
aws s3 ls                                       # list your buckets
```

Many EC2 commands also take `--dry-run`: AWS checks your permissions and parameters, then stops before doing anything. You'll use the same idea from Go in *AWS API errors: EC2 dry-run permission checks*.

```bash
aws ec2 terminate-instances --instance-ids i-0abcd1234ef567890 --dry-run
# An error occurred (DryRunOperation) when calling the TerminateInstances operation:
# Request would have succeeded, but DryRun flag is set.
```

For anything you build more than once, write it down as *infrastructure as code* (Terraform, CloudFormation, CDK): a file describing what should exist, applied by a tool, reviewed like code. It's the same desired-state idea as Kubernetes manifests. The next module, *Cloud automation with the AWS SDK*, does the API side from Go, ending with a tiny version of `terraform plan`.

### CloudTrail: who did that?

*CloudTrail* records API calls in your account: who (which identity and access key), what (`TerminateInstances`), on what, when, from which IP address, and whether it was allowed. Without any setup, the console's *event history* keeps 90 days of calls that create, change or delete things. A *trail* saves events to an S3 bucket for as long as you like; reads of S3 objects (*data events*) are only recorded if you turn them on.

When a key leaks, CloudTrail is how you find out what the attacker did. You'll trace one in *CloudTrail: trace a leaked access key*.

## What it costs

You pay for what you use, along a few *pricing dimensions*:

| Dimension | Example price (us-east-1) | Watch out for |
|---|---|---|
| Instance-hours | `t3.medium`: $0.0416/hour | Forgotten instances running all month |
| Storage per GB-month | EBS gp3: $0.08; S3 Standard: $0.023 | Old snapshots and logs piling up |
| NAT gateway | $0.045/hour + $0.045/GB processed | Heavy traffic through NAT |
| Data transfer out | About $0.09/GB to the internet | Serving big files without a CDN |
| Requests | Per million API or S3 requests | Chatty clients |

Data coming *in* is free; data going *out* costs money, and so does traffic between AZs (a small amount per GB). AWS prices a month as 730 hours (365 × 24 ÷ 12). A quick estimate in Go:

```go
package main

import "fmt"

func main() {
	const hoursPerMonth = 730
	pricePerHour := map[string]float64{"t3.micro": 0.0104, "m5.large": 0.096}

	web := 3 * pricePerHour["m5.large"] * hoursPerMonth
	bastion := pricePerHour["t3.micro"] * hoursPerMonth
	fmt.Printf("web: $%.2f\n", web)
	fmt.Printf("bastion: $%.2f\n", bastion)
	fmt.Printf("total: $%.2f a month\n", web+bastion)
}
```

It prints:

```text
web: $210.24
bastion: $7.59
total: $217.83 a month
```

Three `m5.large` instances cost 3 × $0.096 × 730 = $210.24; one `t3.micro` adds $7.59. Small hourly prices add up because a month has 730 hours.

### Seeing a surprise coming

- **Budgets.** *AWS Budgets* emails you when actual spend, or the forecast for the month, crosses a threshold such as 80% of your budget. Set one on day one of any account.
- **Tags.** Tag resources with `team`, `env` and `service`, and turn those tags on for cost allocation, so the bill can be split by team.
- **Cost Explorer** shows where the money went, by service, account or tag.

You'll build an estimator and budget alerts in *Cloud costs: estimates and budget alarms*.

## Designing for failure

AWS's CTO Werner Vogels sums it up as "everything fails, all the time". The cloud doesn't make failure go away; it gives you cheap ways to survive it:

- **Spread across AZs.** Two instances in two AZs behind a load balancer survive losing a zone. Two in the same AZ don't.
- **Prefer managed failover.** RDS Multi-AZ keeps a standby copy in another AZ and switches to it for you.
- **Replace, don't repair.** An auto scaling group replaces an unhealthy instance with a fresh one, like a Deployment replacing a Pod.
- **Back up across accounts or regions** for data you can't lose, so one mistake or one compromised account can't delete every copy.

## Common mistakes

- **Everything in one account.** One mistake or leaked key reaches everything. Separate accounts per environment.
- **Everything in one AZ.** It's cheaper until the day that zone has a bad hour.
- **Public subnets for databases.** Databases belong in private subnets, reachable only from the app's security group.
- **Opening port 22 to `0.0.0.0/0`.** Use a narrow range, or a session manager instead of SSH.
- **Forgetting NAT and data transfer costs.** They're often the surprise line on the bill.
- **No budget alarm.** You find out about the runaway instance a month later.

## Check yourself

1. A teammate says "I can't see the instance I created" while looking at `us-east-1` in the AWS console. The instance was created by a script configured for `eu-west-1`. What's going on?
2. Why do companies use many AWS accounts instead of one?
3. An app runs two EC2 instances behind a load balancer, both in `eu-west-1a`. What failure does this design not survive, and what would you change?
4. What makes an AWS subnet "public" rather than "private"?
5. A network ACL allows inbound TCP 443 from anywhere and outbound TCP 443 to anywhere, nothing else. HTTPS clients connect but never get a response. Why?
6. Under the shared responsibility model, whose job is it to patch the operating system on an EC2 instance, and whose job is it to keep an S3 bucket private?
7. An `m5.large` costs $0.096 per hour. Roughly what does one cost per month if it runs all the time?

## Answers

1. Regions are independent: a resource only shows up in the region it was created in. Switch the console to `eu-west-1` to see it.
2. The account is AWS's strongest boundary. Separate accounts limit the blast radius of mistakes and leaked keys, keep production apart from experiments, and make the bill easy to split.
3. Losing the AZ `eu-west-1a` takes both instances down. Put the instances in subnets in at least two different AZs.
4. Its route table: a public subnet sends `0.0.0.0/0` to an internet gateway; a private one sends it to a NAT gateway or nowhere.
5. Network ACLs are stateless. Replies go back to the client's ephemeral port (1024–65535), not to port 443, and the outbound rules don't allow those ports, so every reply is dropped.
6. Patching the EC2 operating system is the customer's job (security *in* the cloud). Keeping the bucket private is also the customer's job: AWS stores the bytes safely, but the bucket policy and access settings are yours.
7. About $70: 0.096 × 730 hours = $70.08.

## Key terms

| Term | Meaning |
|---|---|
| Public cloud | Computing resources rented on demand through an API, billed by use |
| Shared responsibility model | AWS secures the cloud itself; the customer secures what they put in it |
| AWS account | A container for resources with its own bill and permissions, identified by a 12-digit ID |
| AWS Organizations | A service that groups many accounts, with one bill and shared rules |
| Region | A geographic area with its own independent copy of AWS services, like `eu-west-1` |
| Availability zone | One or more isolated data centres inside a region, like `eu-west-1a` |
| ARN | Amazon Resource Name: `arn:partition:service:region:account:resource` |
| EC2 | AWS's virtual machines |
| S3 | AWS's object storage: objects stored under keys in buckets |
| VPC | A private network in one region with an address range you choose |
| Subnet | A slice of a VPC's address range, in exactly one availability zone |
| Route table | The rules that decide where packets leaving a subnet go next |
| Internet gateway | The VPC's door to the internet, used by public subnets |
| NAT gateway | Lets instances in private subnets reach out without being reachable |
| Security group | A stateful, allow-only firewall on an instance's network interface |
| Network ACL | A stateless subnet firewall with numbered allow and deny rules |
| CloudTrail | AWS's record of API calls: who did what, when, from where |
| Pricing dimension | A unit AWS bills by, such as instance-hours or GB-months |
| AWS Budgets | Alerts when actual or forecast spend crosses a threshold |
| Infrastructure as code | Describing infrastructure in reviewed files that a tool applies |

## Watch

- [AWS regions, availability zones and why you should care](https://www.youtube.com/watch?v=LPwYyTee6oE) · Dmitry Figol · 14 min: what an AZ and a region are, measured latency between zones, and why your architecture should care.
- [AWS re:Invent 2016: Tuesday Night Live with James Hamilton](https://www.youtube.com/watch?v=AyOAjFNPAbA) · Amazon Web Services · 93 min: a classic tour of what's physically behind a region and an AZ, from undersea cables to power substations; watch Hamilton's own segments and skip the customer guests.
- [AWS re:Invent 2025 - AWS Networking Fundamentals: Connect, secure and scale (NET208)](https://www.youtube.com/watch?v=nXBqxnp5ybY) · AWS Events · 59 min: starts from one VPC in one region (addressing, subnets, routing, DNS, security) before growing to many; the early part matches this guide.

## Go deeper

- [Overview of Amazon Web Services](https://docs.aws.amazon.com/whitepapers/latest/aws-overview/introduction.html) (AWS): the official tour of the main services and global infrastructure.
- [AWS Regions and Availability Zones](https://docs.aws.amazon.com/global-infrastructure/latest/regions/aws-regions-availability-zones.html) (AWS): how regions, zones and AZ IDs work.
- [How Amazon VPC works](https://docs.aws.amazon.com/vpc/latest/userguide/how-it-works.html) (AWS): subnets, route tables and gateways with diagrams.
- [Shared Responsibility Model](https://aws.amazon.com/compliance/shared-responsibility-model/) (AWS): exactly which security jobs are AWS's and which are yours.
- [Managing your costs with AWS Budgets](https://docs.aws.amazon.com/cost-management/latest/userguide/budgets-managing-costs.html) (AWS): setting budgets and actual or forecasted alerts.
