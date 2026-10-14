## Task 1: Name the severities

Choose a word based on the value of `s`, with a fallback for numbers that
have no name. A `switch` with one case per constant fits well.

### Closer

Use `switch s` with `case Minor:`, `case Major:` and `case Critical:`, plus a
`default` that formats the number with `%d` and `int(s)`. Add
`import "fmt"` under the `package` line.

### Answer

```go
switch s {
case Minor:
	return "minor"
case Major:
	return "major"
case Critical:
	return "critical"
default:
	return fmt.Sprintf("Severity(%d)", int(s))
}
```

`int(s)` has no `String` method, so `%d` prints the plain number.

## Task 2: Print an incident

The only part that changes shape is inside the parentheses: the state, then
maybe ` by ` and the owner. Work that part out first.

### Closer

Start with `status := string(i.State)`, add `" by " + i.Owner` when `Owner`
isn't empty, then use `fmt.Sprintf("#%d [%v] %s (%s)", ...)`. `%v` on
`i.Severity` calls your `Severity.String`.

### Answer

```go
status := string(i.State)
if i.Owner != "" {
	status += " by " + i.Owner
}
return fmt.Sprintf("#%d [%v] %s (%s)", i.ID, i.Severity, i.Title, status)
```

## Task 3: Acknowledge and resolve

Each method first asks "is there anything to do in this state?" and returns
`false` straight away if not. Only then does it change fields.

### Closer

`Ack` returns early with `if i.State != StateOpen { return false }`, then
sets `i.State` and `i.Owner`. `Resolve` returns early when `i.State ==
StateResolved`. With a pointer receiver, `i.State = ...` changes the real
incident; no `*` needed.

### Answer

```go
func (i *Incident) Ack(owner string) bool {
	if i.State != StateOpen {
		return false
	}
	i.State = StateAcked
	i.Owner = owner
	return true
}

func (i *Incident) Resolve() bool {
	if i.State == StateResolved {
		return false
	}
	i.State = StateResolved
	return true
}
```

## Task 4: Open and find incidents

`Open` bumps `lastID` and appends a new `Incident` built from it. `Find`
loops over the board and returns the address of the match, not of a copy.

### Closer

In `Open`: `b.lastID++`, then `b.incidents = append(b.incidents,
Incident{ID: b.lastID, ...})`. In `Find`: `for i := range b.incidents`, and
when `b.incidents[i].ID == id`, return `&b.incidents[i]`. After the loop,
return `nil`.

### Answer

```go
func (b *Board) Open(title string, sev Severity) int {
	b.lastID++
	b.incidents = append(b.incidents, Incident{
		ID:       b.lastID,
		Title:    title,
		Severity: sev,
		State:    StateOpen,
	})
	return b.lastID
}

func (b *Board) Find(id int) *Incident {
	for i := range b.incidents {
		if b.incidents[i].ID == id {
			return &b.incidents[i]
		}
	}
	return nil
}
```

## Task 5: Pick the worst open incident

Walk the board once, remembering the worst unresolved incident so far.
Replace it only when you find a strictly higher severity, so ties keep the
earlier one.

### Closer

Start with `var worst *Incident`. Loop with `for i := range b.incidents`,
take `inc := &b.incidents[i]`, skip resolved ones with `continue`, and set
`worst = inc` when `worst == nil || inc.Severity > worst.Severity`.

### Answer

```go
var worst *Incident
for i := range b.incidents {
	inc := &b.incidents[i]
	if inc.State == StateResolved {
		continue
	}
	if worst == nil || inc.Severity > worst.Severity {
		worst = inc
	}
}
return worst
```

`>` rather than `>=` keeps the first incident on a tie.
