# Cloud costs: estimates and budget alarms

In the cloud, every server, disk and gigabyte has a meter running. Platform
engineers are expected to say what a design will cost *before* it is built,
and to hear about a runaway bill on day 10, not on the invoice. Here you
write a small cost estimator and the alarm logic of a monthly budget.

**New here:** pricing dimensions, AWS Budgets (actual and forecasted alerts), `math.Round`.

## Learn

### Pay for what you use

AWS has no monthly plans for most services. Each service is priced along
one or more *pricing dimensions*, and you pay quantity × unit price:

| Service | You pay for | Unit |
|---|---|---|
| EC2 (servers) | each hour an instance runs | instance-hour, by instance type |
| EBS (disks) | the disk size, all month, used or not | GB-month |
| S3 (object storage) | what you store | GB-month |
| NAT gateway | each hour it exists, **plus** every GB through it | hour + GB |
| Data transfer out | data sent from AWS to the internet | GB |

A month is billed as **730 hours** (365 days × 24 hours ÷ 12). An instance
that runs all month costs 730 × its hourly price; the starter's
`HoursPerMonth` holds 730.

### Real prices

These are on-demand prices in `us-east-1` (US East, N. Virginia) for Linux,
in US dollars. They are in `testdata/prices.json`, and the starter's
`LoadPrices` reads that file into a `Prices` struct.

| What | Price | Per month, nonstop |
|---|---|---|
| EC2 `t3.micro` (2 vCPU, 1 GiB) | $0.0104 per hour | $7.59 |
| EC2 `t3.small` (2 vCPU, 2 GiB) | $0.0208 per hour | $15.18 |
| EC2 `t3.medium` (2 vCPU, 4 GiB) | $0.0416 per hour | $30.37 |
| EC2 `m5.large` (2 vCPU, 8 GiB) | $0.096 per hour | $70.08 |
| EC2 `c5.xlarge` (4 vCPU, 8 GiB) | $0.17 per hour | $124.10 |
| EBS `gp3` disk | $0.08 per GB-month | |
| S3 Standard | $0.023 per GB-month | |
| NAT gateway | $0.045 per hour + $0.045 per GB | $32.85 before any traffic |
| Data transfer out to the internet | $0.09 per GB | |

Other regions cost a bit more or less, and prices change; the AWS pricing
pages are the source of truth.

### Where surprise bills come from

- **NAT gateways.** One per availability zone, as *VPC planning: subnets per availability zone* suggests, costs about $99 a month for three zones before a byte flows.
- **Data transfer out.** Serving 10 TB of downloads is about $900.
- **Forgotten resources.** A test `c5.xlarge` left running is $124 a month;
  its disk keeps costing after the instance stops.
- **The free tier ends.** New accounts get some usage free for a while;
  check what yours covers and when it stops.

*Cost allocation tags* help find the owner: tag resources with keys such as
`CostCenter`, activate those keys in the Billing console, and the bill can
be split by them.

### AWS Budgets

A *budget* in AWS Budgets is a monthly amount, say $1000, with alerts at
percentages of it. Each alert is one of two kinds:

| Kind | Fires when | Tells you |
|---|---|---|
| actual | what you spent so far reaches X% of the budget | it already happened |
| forecasted | the predicted month-end total reaches X% | it will happen if nothing changes |

AWS predicts with your history. This exercise uses a straight line: if you
spent $300 in the first 10 days of a 30-day month, the forecast is
$300 / 10 × 30 = $900.

### Floats and rounding to cents

Money here is a `float64`. Floats can't store most decimals exactly, so
sums pick up tiny errors: `0.1 + 0.2` is `0.30000000000000004`. Round
results to cents with `math.Round`, which rounds to the nearest whole
number (halves away from zero):

```go
math.Round(2.5)              // 3
math.Round(17.4999)          // 17
math.Round(3.14159*1000) / 1000 // 3.142: three decimals
```

Rounding once at the end keeps the errors small. Real billing systems
avoid floats altogether and count whole cents in an integer.

> **Why is it written like this?** `math.Round` only rounds to whole
> numbers, so you shift the decimal point, round, and shift back.
> Multiplying and dividing by a power of ten is the usual idiom.

### Try it (optional)

These commands only read. Needs an AWS account and credentials; never
needed for the tests.

```bash
aws ce get-cost-and-usage --time-period Start=2026-09-01,End=2026-10-01 --granularity MONTHLY --metrics UnblendedCost --group-by Type=DIMENSION,Key=SERVICE   # last month's bill per service
aws budgets describe-budgets --account-id 111122223333   # your budgets and their limits
aws pricing get-products --region us-east-1 --service-code AmazonEC2 --filters Type=TERM_MATCH,Field=instanceType,Value=t3.micro Type=TERM_MATCH,Field=regionCode,Value=us-east-1 --max-results 1   # AWS's own price list entry
```

Use your own account ID. Each Cost Explorer API call costs $0.01.

## Your task

Your team wants a cost estimate for a new service, and a budget that warns
before the month ends. The tests load `testdata/prices.json` with
`LoadPrices`.

### 1. Price one line

**Goal:** `LineCost(item, p)` returns the item's monthly cost in dollars, **not** rounded.

`item.Service` decides the formula. Service names are lowercase, exactly as
below:

| Service | Cost |
|---|---|
| `"ec2"` | `Count` × `Hours` × `p.EC2Hourly[InstanceType]` |
| `"ebs"` | `GB` × `p.EBSPerGB` |
| `"s3"` | `GB` × `p.S3PerGB` |
| `"nat"` | `Count` × `Hours` × `p.NATHourly` + `GB` × `p.NATPerGB` |
| `"transfer"` | `GB` × `p.TransferPerGB` |

Return `0` and an error for any other service (for example
`unknown service "lambda"`), and for an `"ec2"` item whose `InstanceType`
has no price (for example `no price for instance type "t3.nano"`).

```go
LineCost(Item{Service: "ec2", InstanceType: "t3.micro", Count: 2, Hours: HoursPerMonth}, p)
// => 15.184, nil
LineCost(Item{Service: "nat", Count: 1, Hours: HoursPerMonth, GB: 250}, p)
// => 44.1, nil  (32.85 for the hours + 11.25 for the traffic)
LineCost(Item{Service: "s3", GB: 1200}, p) // => 27.6, nil
LineCost(Item{Service: "lambda", GB: 10}, p) // => 0, error
```

**Think it through:**
- `Count` is an `int` and the prices are `float64`. What must you do before multiplying them?
- How does a two-value map lookup tell you an instance type has no price?

### 2. Estimate the bill

**Goal:** `Estimate(items, p)` returns a `Bill` with the total and the cost per service, each rounded to cents.

- Add up the **unrounded** `LineCost` of every item, both into the total
  and into `ByService[item.Service]`.
- Then round the total and each service's sum to cents, once, at the end.
- If any `LineCost` fails, return `Bill{}` and that error.
- No items: `Total` is `0` and `ByService` is empty.

```go
Estimate([]Item{
	{Service: "ec2", InstanceType: "t3.small", Count: 2, Hours: HoursPerMonth},
	{Service: "ec2", InstanceType: "t3.micro", Count: 1, Hours: HoursPerMonth},
	{Service: "ebs", GB: 60},
	{Service: "s3", GB: 120},
	{Service: "nat", Count: 1, Hours: HoursPerMonth, GB: 80},
	{Service: "transfer", GB: 150},
}, p)
// => Bill{Total: 95.47, ByService: {ec2: 37.96, ebs: 4.8, s3: 2.76, nat: 36.45, transfer: 13.5}}
```

Rounding once at the end matters. One `t3.micro` hour ($0.0104), 1 GB in S3
($0.023) and 0.05 GB out ($0.0045) total `0.04`, with `ec2: 0.01`,
`s3: 0.02` and `transfer: 0`. Rounding each line first would give a total
of `0.03`.

**Think it through:**
- What does `ByService` have to be before you can add to it?
- When do you loop over the map to round its values: during the item loop, or after?

### 3. Raise budget alerts

**Goal:** `BudgetAlerts(budget, spent, day, days, thresholds)` returns the alerts that fire, in `thresholds` order.

- `spent` is what was spent in the first `day` days of a `days`-day month
  (`day` is at least 1). The forecast is `spent / day × days`.
- For each threshold `t` (a percentage, given in ascending order), the
  limit is `budget × t / 100`:
  - `spent` at or above the limit: `ACTUAL <t>%: $<spent> of $<budget>`;
  - otherwise, forecast at or above the limit:
    `FORECASTED <t>%: $<forecast> of $<budget>`;
  - otherwise no alert for `t`.
- Dollar amounts have two decimals. No alerts: an empty result (`nil` is
  fine).

```go
BudgetAlerts(1000, 812.50, 20, 30, []int{50, 80, 100, 120})
// => ["ACTUAL 50%: $812.50 of $1000.00"
//     "ACTUAL 80%: $812.50 of $1000.00"
//     "FORECASTED 100%: $1218.75 of $1000.00"
//     "FORECASTED 120%: $1218.75 of $1000.00"]
BudgetAlerts(200, 160, 30, 30, []int{80, 100})
// => ["ACTUAL 80%: $160.00 of $200.00"]  (exactly at 80% counts)
BudgetAlerts(1000, 100, 10, 30, []int{50, 80, 100})
// => []  (forecast $300 reaches no threshold)
```

**Think it through:**
- `day`, `days` and `t` are `int`s, the money is `float64`. Where do you convert?
- Which `fmt.Sprintf` verbs print a whole number, a float with two decimals, and a literal `%`?

## Learn more

- [Managing your costs with AWS Budgets](https://docs.aws.amazon.com/cost-management/latest/userguide/budgets-managing-costs.html)
- [Organizing and tracking costs using AWS cost allocation tags](https://docs.aws.amazon.com/awsaccountbilling/latest/aboutv2/cost-alloc-tags.html)
- [Pricing for NAT gateways](https://docs.aws.amazon.com/vpc/latest/userguide/nat-gateway-pricing.html)
- [math.Round](https://pkg.go.dev/math#Round)
