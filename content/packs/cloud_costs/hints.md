## Task 1: Price one line

A `switch` on `item.Service` with one `case` per service from the table,
each returning its formula. Whatever falls through the switch is an unknown
service.

### Closer

Convert the count before multiplying: `float64(item.Count) * item.Hours`.
For EC2, look the price up with `hourly, ok := p.EC2Hourly[item.InstanceType]`
and return an error when `!ok`. Errors come from `fmt.Errorf("unknown service %q", item.Service)`.

### Answer

```go
switch item.Service {
case "ec2":
	hourly, ok := p.EC2Hourly[item.InstanceType]
	if !ok {
		return 0, fmt.Errorf("no price for instance type %q", item.InstanceType)
	}
	return float64(item.Count) * item.Hours * hourly, nil
case "ebs":
	return item.GB * p.EBSPerGB, nil
case "s3":
	return item.GB * p.S3PerGB, nil
case "nat":
	return float64(item.Count)*item.Hours*p.NATHourly + item.GB*p.NATPerGB, nil
case "transfer":
	return item.GB * p.TransferPerGB, nil
}
return 0, fmt.Errorf("unknown service %q", item.Service)
```

## Task 2: Estimate the bill

Make the map first, then loop over the items adding each `LineCost` to two
places. Rounding comes after the loop, for the total and for every map
value.

### Closer

`bill := Bill{ByService: map[string]float64{}}`, then
`bill.ByService[item.Service] += cost`. A small helper
`math.Round(dollars*100) / 100` rounds to cents; add `"math"` to the
imports. Writing to existing keys while ranging over a map is allowed.

### Answer

```go
bill := Bill{ByService: map[string]float64{}}
for _, item := range items {
	cost, err := LineCost(item, p)
	if err != nil {
		return Bill{}, err
	}
	bill.Total += cost
	bill.ByService[item.Service] += cost
}
bill.Total = math.Round(bill.Total*100) / 100
for service, cost := range bill.ByService {
	bill.ByService[service] = math.Round(cost*100) / 100
}
return bill, nil
```

## Task 3: Raise budget alerts

Work out the forecast once, before the loop. Then, for each threshold,
compare `spent` with the limit first and the forecast only if that didn't
fire.

### Closer

`forecast := spent / float64(day) * float64(days)` and
`limit := budget * float64(t) / 100`. Use `if` / `else if` for the two
alerts, and `fmt.Sprintf("ACTUAL %d%%: $%.2f of $%.2f", t, spent, budget)`.

### Answer

```go
forecast := spent / float64(day) * float64(days)
var alerts []string
for _, t := range thresholds {
	limit := budget * float64(t) / 100
	if spent >= limit {
		alerts = append(alerts, fmt.Sprintf("ACTUAL %d%%: $%.2f of $%.2f", t, spent, budget))
	} else if forecast >= limit {
		alerts = append(alerts, fmt.Sprintf("FORECASTED %d%%: $%.2f of $%.2f", t, forecast, budget))
	}
}
return alerts
```
