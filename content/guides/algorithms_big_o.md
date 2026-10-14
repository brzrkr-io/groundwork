# Algorithms and Big-O

**In one sentence:** Big-O describes how the work an algorithm does grows as
its input grows, so you can tell which code will still be fast when the log
file is a thousand times bigger.

**Why it matters:** a script that checks 1,000 hosts in a blink can take
*days* on 1,000,000 hosts if it compares every host with every other one.
The fix is often a three-line change, but only if you can spot the problem.

**You'll learn:**

- What an algorithm is, and why we count steps instead of seconds.
- The five growth rates you'll meet most: O(1), O(log n), O(n),
  O(n log n) and O(n²), with real numbers.
- Linear search versus binary search, and why binary search needs sorted
  data.
- What sorting costs, and how to escape the nested-loop trap with a map.
- How to trade memory for speed, and when none of this matters.

## What an algorithm is

An *algorithm* is a precise, step-by-step recipe for solving a problem. "Find
the 503 errors in this log" is a problem. "Read each line; if it contains
` 503 `, print it" is an algorithm for it.

A program is an algorithm written in a language a computer can run. The same
algorithm can be written in Go, Python or Bash; the recipe is the same.

Most problems have several algorithms. They all give the right answer, but
some do far less work. This guide is about telling them apart *before* you
run them on production data.

## Counting steps, not seconds

You could time your code with a stopwatch, but seconds depend on the laptop,
on what else is running, and on the input you happened to try. So computer
scientists ask a different question:

> When the input doubles, what happens to the amount of work?

Count the basic *steps*: comparisons, additions, map lookups. Then see how
that count changes as the input size grows. We call the input size *n*: the
number of log lines, hosts, or pods.

| Task on n log lines | Steps | When n doubles... |
|---|---|---|
| Read the first line | 1 | still 1 |
| Count every line | n | twice the work |
| Compare every line with every other | about n × n / 2 | four times the work |

*Big-O notation* is the shorthand for that last column. It keeps only the
part that grows fastest and drops constant factors:

- n × n / 2 is written O(n²): the "/ 2" doesn't change the shape.
- 3n + 7 is written O(n): when n is a million, the 7 is noise, and the 3 is
  a speed difference, not a growth difference.

Big-O usually describes the *worst case*: the most work the algorithm can be
forced to do. Searching a list is worst when the item is last or missing.

## The five growth rates you'll meet

| Big-O | Name | Ops example |
|---|---|---|
| O(1) | constant | Look up a host in a map; read `s[5]` |
| O(log n) | logarithmic | Binary search a sorted host list |
| O(n) | linear | Scan every line of a log once |
| O(n log n) | "n log n" | Sort the log lines by timestamp |
| O(n²) | quadratic | Compare every host with every other host |

*log n* here means "how many times can you halve n until you reach 1?".
Halving 1,000 takes about 10 halvings; halving 1,000,000 takes about 20. It
grows incredibly slowly.

Now the numbers. Say one step takes 1 microsecond, about the time to parse a
short log line.

| Big-O | Steps, 1,000 lines | Steps, 1,000,000 lines | Time, 1,000,000 |
|---|---|---|---|
| O(1) | 1 | 1 | 1 µs |
| O(log n) | 10 | 20 | 20 µs |
| O(n) | 1,000 | 1,000,000 | 1 second |
| O(n log n) | 10,000 | 20,000,000 | 20 seconds |
| O(n²) | 1,000,000 | 1,000,000,000,000 | about 11.6 days |

Read the last row twice. On 1,000 lines, the O(n²) script takes one second
and passes every test. On a real day's logs it would still be running after
the next sprint. **Growth rate beats raw speed**: no faster laptop rescues
an O(n²) algorithm on big input.

## Linear search: check every item

The obvious way to find something in a slice is to look at each item until
you find it. That's *linear search*, and it's O(n): in the worst case you
look at all n items.

It needs nothing special from the data: any order works. For a slice of 50
hosts it's the right choice. For a million hosts, searched thousands of
times, it isn't.

## Binary search: halve the problem

Think of finding a name in a printed phone book. You don't start at page 1.
You open the middle: if the name you want comes later, the whole first half
is gone. Repeat with the half that's left.

```text
  looking for "host-0999999" in 1,000,000 sorted hosts

  step 1: [..................|..................]  middle too small,
                                                   keep the right half
  step 2:                    [........|........]   keep the right half
  step 3:                              [...|...]   keep the right half
  ...
  step 20:                                   [x]   found it
```

That's *binary search*. Each step throws away half of what's left, so a
million items take about 20 steps: O(log n).

**The catch: the data must be sorted.** "The middle is too small, so go
right" only works if everything to the right is bigger.

Here are both searches side by side, counting their steps:

```go
package main

import "fmt"

// linearSearch checks every host in turn and counts the comparisons.
func linearSearch(hosts []string, target string) (int, int) {
	steps := 0
	for i, h := range hosts {
		steps++
		if h == target {
			return i, steps
		}
	}
	return -1, steps
}

// binarySearch needs hosts sorted. It halves the range on every step.
func binarySearch(hosts []string, target string) (int, int) {
	steps := 0
	lo, hi := 0, len(hosts)-1
	for lo <= hi {
		steps++
		mid := lo + (hi-lo)/2
		switch {
		case hosts[mid] == target:
			return mid, steps
		case hosts[mid] < target:
			lo = mid + 1 // target is in the right half
		default:
			hi = mid - 1 // target is in the left half
		}
	}
	return -1, steps
}

func main() {
	const n = 1000000
	hosts := make([]string, 0, n)
	for i := range n {
		hosts = append(hosts, fmt.Sprintf("host-%07d", i))
	}
	target := "host-0999999"

	i, steps := linearSearch(hosts, target)
	fmt.Println("linear: found at", i, "after", steps, "steps")

	i, steps = binarySearch(hosts, target)
	fmt.Println("binary: found at", i, "after", steps, "steps")
}
```

It prints:

```text
linear: found at 999999 after 1000000 steps
binary: found at 999999 after 20 steps
```

Trace it:

1. `main` builds a million names, `host-0000000` to `host-0999999`. The
   `%07d` pads numbers with zeros, so alphabetical order matches number
   order: the slice is already sorted.
2. `linearSearch` starts at index 0 and only finds the target on its last
   look: 1,000,000 steps.
3. `binarySearch` keeps two edges, `lo` and `hi`. Each loop looks at the
   middle and moves one edge past it, halving the range. After 20 halvings
   it lands on the target.

You don't have to write binary search yourself: `slices.BinarySearch` does
it for you, as the next section shows.

## Sorting: paying once to search fast

Sorting puts items in order: hosts by name, alerts by count, log lines by
time. Go's `slices.Sort` and `slices.SortFunc` do it for you; you'll use them in
*Sorting*. Older code
uses `sort.Slice` from the `sort` package; it does the same job.

```go
package main

import (
	"fmt"
	"slices"
)

func main() {
	hosts := []string{"web-3", "db-1", "web-1", "cache-2"}
	slices.Sort(hosts)
	fmt.Println(hosts)

	i, found := slices.BinarySearch(hosts, "web-1")
	fmt.Println(i, found)

	i, found = slices.BinarySearch(hosts, "web-2")
	fmt.Println(i, found)
}
```

It prints:

```text
[cache-2 db-1 web-1 web-3]
2 true
3 false
```

`web-1` is at index 2. `web-2` isn't there, so `found` is `false`, and 3 is
where it *would* go to keep the slice sorted.

**How much does sorting cost?** Go's sort does O(n log n) comparisons. The
intuition: about log n rounds of splitting the data in half, and each round
touches all n items. Mathematicians have proved that any sort which works by
comparing pairs needs about n log n comparisons in the worst case, so O(n
log n) is as good as it gets.

**When is sorting worth it?** One search: just scan, O(n). Thousands of
searches on the same data: sort once for O(n log n), then each search is
O(log n). If you only ever ask "is X there?", a map is simpler still: see the
*Data structures* guide.

## The nested-loop trap

Here's the most common performance bug in ops scripts. You have an inventory
and you want to find hosts listed twice. The obvious idea: compare every host
with every other host.

```text
  for each host A:
      for each host B after A:
          if A == B: report a duplicate
```

A loop inside a loop over the same data is the classic *O(n²)* shape. With n
hosts it makes about n × n / 2 comparisons.

**The fix: remember what you've seen.** One pass over the hosts with a map
answers "seen it before?" in O(1) each time, so the whole job is O(n).

```go
package main

import "fmt"

// dupesSlow compares every host with every host after it.
func dupesSlow(hosts []string) ([]string, int) {
	var dupes []string
	steps := 0
	for i := 0; i < len(hosts); i++ {
		for j := i + 1; j < len(hosts); j++ {
			steps++
			if hosts[i] == hosts[j] {
				dupes = append(dupes, hosts[i])
			}
		}
	}
	return dupes, steps
}

// dupesFast remembers every host it has seen in a map.
func dupesFast(hosts []string) ([]string, int) {
	var dupes []string
	steps := 0
	seen := make(map[string]bool, len(hosts))
	for _, h := range hosts {
		steps++
		if seen[h] {
			dupes = append(dupes, h)
		}
		seen[h] = true
	}
	return dupes, steps
}

func main() {
	var hosts []string
	for i := range 10000 {
		hosts = append(hosts, fmt.Sprintf("node-%d", i))
	}
	hosts = append(hosts, "node-42") // someone pasted node-42 twice

	d, steps := dupesSlow(hosts)
	fmt.Println("slow:", d, "in", steps, "steps")

	d, steps = dupesFast(hosts)
	fmt.Println("fast:", d, "in", steps, "steps")
}
```

It prints:

```text
slow: [node-42] in 50005000 steps
fast: [node-42] in 10001 steps
```

Same answer. With 10,001 hosts, the nested loop does 50 million comparisons;
the map does 10,001 lookups. Double the hosts and the slow version does four
times the work, while the fast one does twice.

### Hidden nested loops

The trap isn't always two `for` lines you can see. A function that searches
a slice is a loop too:

```text
  for each of 100,000 pods:
      if slices.Contains(allowedImages, pod.image)   <- loops over
          ...                                           allowedImages
```

`slices.Contains` is O(m) for a list of m items, so this is O(n × m). Same
fix: build a set (`map[string]bool`) from `allowedImages` once, before the
loop.

## Space versus time

The map in `dupesFast` wasn't free: it holds every host name in memory. We
spent *space* (memory) to save *time*. That's one of the most common trades
in computing.

| Trade | Spend | Save |
|---|---|---|
| Build a map for lookups | memory for the map | a loop per lookup |
| Sort once up front | one O(n log n) sort | many O(n) scans |
| Cache an API response | memory, stale data risk | network round trips |
| Stream a log line by line | re-reading if needed again | holding 20 GB in RAM |

The last row goes the other way. Reading a huge log one line at a time
uses a tiny, fixed amount of memory, O(1) space, where loading every line
into a slice first needs O(n). You'll do that in the *Text and data* module.

Big-O applies to memory too: "this function uses O(n) extra space" means its
memory grows in step with its input.

## When Big-O doesn't matter

Big-O tells you how work *grows*. It doesn't tell you which code is faster
on the data you actually have.

- **Small n.** Scanning a slice of 20 regions is effectively instant. A map
  would be no faster and less clear. Don't "optimise" this.
- **Something else dominates.** One HTTP call to an API takes around
  50 milliseconds; a map lookup takes a few nanoseconds. If your tool makes
  200 API calls, the loops around them don't matter. Cut the calls.
- **Constants are real.** O(n) code that does a slow thing per item (opens
  a file, calls a server) can be slower than O(n log n) code that does a
  cheap one, for any input you'll ever see.

**Measure first.** When something is slow, time it on realistic data before
changing code. You'll learn Go's built-in tool for this in *Benchmarks*, in
the *Testing* module. Then use Big-O to explain what you measured and to
predict what happens at ten times the size.

A good habit: when you write a loop, ask "how big can this get in
production?" If the answer is "every log line" or "every pod in the
cluster", look for loops inside it.

## Common mistakes

- **Testing only on tiny data.** O(n²) code passes on a 10-line fixture and
  falls over on a real log. Try one big input before you ship.
- **Binary searching unsorted data.** It doesn't crash; it quietly returns
  wrong answers. Sort first, or use a map.
- **Sorting just to find one thing.** A single O(n) scan beats an
  O(n log n) sort followed by a search.
- **Missing the hidden loop.** `slices.Contains`, `slices.Index` and
  `strings.Contains` all loop. Calling them inside another loop multiplies.
- **Optimising without measuring.** Shaving loops in a script that spends
  99% of its time waiting on the network changes nothing.
- **Thinking Big-O is the running time.** O(n) says how the work grows, not
  how many seconds it takes.

## Check yourself

1. A function does `3n + 100` steps. What is its Big-O, and why can you drop
   the 3 and the 100?
2. Your script takes 2 seconds on a 10,000-line log and is O(n²). Roughly how
   long on a 100,000-line log?
3. You must check 50,000 incoming IPs against a blocklist of 10,000 IPs.
   What's slow about looping over the blocklist for each IP, and what's the
   fix?
4. Why does binary search need sorted data?
5. You search a list of 12 Kubernetes namespaces once per run. Should you
   sort it and binary search, or build a map?
6. Building a `seen` map makes duplicate detection O(n). What did it cost?

## Answers

1. O(n). Big-O describes how work grows. As n gets large, the 100 is
   negligible, and the factor 3 changes the speed but not the shape: doubling
   n still doubles the work.
2. About 200 seconds. The input is 10 times bigger, and O(n²) means
   10 × 10 = 100 times the work: 2 s × 100.
3. It's O(n × m): 50,000 × 10,000 = 500 million comparisons. Put the
   blocklist into a `map[string]bool` once (10,000 steps), then check each
   IP with one lookup (50,000 steps).
4. Each step looks at the middle and throws away the half that "can't"
   contain the target. That reasoning only holds if items are in order;
   on unsorted data it throws away the wrong half and misses items.
5. Neither. With 12 items, a plain loop or `slices.Contains` is instant and
   simplest. Big-O matters when n is big.
6. Memory: the map stores every host name. That's the space-for-time trade.

## Key terms

| Term | Meaning |
|---|---|
| algorithm | A precise, step-by-step recipe for solving a problem. |
| n | The size of the input: lines, hosts, pods. |
| Big-O | How work grows with n, ignoring constants and smaller terms. |
| worst case | The most work the algorithm can be made to do. |
| O(1), constant | Same work no matter how big n is. |
| O(log n), logarithmic | Work grows by one step each time n doubles. |
| O(n), linear | Work grows in step with n. |
| O(n log n) | The cost of a good comparison sort. |
| O(n²), quadratic | Double n, four times the work; nested loops. |
| linear search | Check items one by one until you find the target. |
| binary search | Halve a sorted range each step until you find the target. |
| space vs time | Using more memory to do less work, or the reverse. |

## Watch

- [Why My Teenage Code Was Terrible: Sorting Algorithms and Big O Notation](https://www.youtube.com/watch?v=RGuJga2Gl_k) · Tom Scott · 10 min: a real story of code that worked on small data and crawled on big data, exactly *The nested-loop trap*.
- [Big O Notation](https://www.youtube.com/watch?v=v4cd1O4zkGw) · HackerRank · 9 min: a clear walk-through of dropping constants, dropping smaller terms and the common growth rates.
- [CS50x 2025 - Lecture 3 - Algorithms](https://www.youtube.com/watch?v=iCx3zwK8Ms8) · CS50 · 127 min: Harvard's famous lecture, with linear versus binary search and a live race between sorting algorithms; watch the searching part first, the sorting part when you have time.

## Go deeper

- [Big O notation (Wikipedia)](https://en.wikipedia.org/wiki/Big_O_notation):
  the formal definition and a table of common orders, once the intuition is
  solid.
- [Binary search (Wikipedia)](https://en.wikipedia.org/wiki/Binary_search):
  history, variations and the classic off-by-one bugs in implementations.
- [Package slices](https://pkg.go.dev/slices): `Sort`, `BinarySearch`,
  `Contains` and the rest of the toolbox used in this guide.
- [Go Slices: usage and internals](https://go.dev/blog/slices-intro): how a
  slice grows its backing array, which explains what `append` really costs.
