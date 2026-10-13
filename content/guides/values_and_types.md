# Values and types

**In one sentence:** everything a program handles is stored as bits, and a
value's type tells Go what those bits mean and what you're allowed to do
with them.

**Why it matters:** ops work is full of values that look alike but aren't.
The port `8080` and the text `"8080"`, 512 MiB and 512 MB, a disk at 45% and
a disk at 0% because of integer division. Types are how Go stops you mixing
them up before your tool reaches production.

**You'll learn:**

- What bits and bytes are, and how numbers are written in binary.
- Why the same bytes can mean different things, and how types settle it.
- The difference between `"8080"`, `8080` and `port` (literals vs names).
- How `int` and `float64` behave: integer division, overflow and rounding.
- That a string is a row of bytes, and what UTF-8 means for its length.
- Why Go makes you convert types yourself, and what zero values are.

## Everything is bits

### Bits and bytes

A computer stores everything as *bits*. A bit is one switch that is either
off or on, written `0` or `1`. One bit can't say much, so bits are grouped in
eights. Eight bits make a *byte*.

```text
one byte:   0 1 0 0 0 0 0 1
            │             │
            │             └ the "1s" place
            └ the "128s" place
```

Memory sizes, file sizes and network traffic are all counted in bytes.

### Counting in binary

You count in tens: each place is worth ten times the one to its right.
*Binary* counts in twos: each place is worth twice the one to its right.

| Binary | Sum of the places that are on | Number |
|---|---|---|
| `0000 0000` | nothing | 0 |
| `0000 0001` | 1 | 1 |
| `0000 0101` | 4 + 1 | 5 |
| `0100 0001` | 64 + 1 | 65 |
| `1111 1111` | 128 + 64 + 32 + 16 + 8 + 4 + 2 + 1 | 255 |

So one byte holds 256 different values: 0 to 255. Each extra bit doubles the
count:

| Bits | Different values | Where you meet it |
|---|---|---|
| 8 | 256 | One byte |
| 16 | 65,536 | Port numbers: 0 to 65,535 |
| 32 | about 4.3 billion | Every IPv4 address there is |
| 64 | about 18 quintillion | Go's `int` on today's machines |

That's why no port is above 65,535: the network sends a port in 16 bits.

### KB, MB, MiB and GiB

Bytes come in two families of bigger units, and ops work uses both.

| Unit | Bytes | Used by |
|---|---|---|
| KB, MB, GB | 1,000 / 1,000,000 / 1,000,000,000 | Disk makers, network speeds |
| KiB, MiB, GiB | 1,024 / 1,048,576 / 1,073,741,824 | RAM, most OS tools |

They drift apart as they grow. A "500 GB" disk is about 465 GiB, so `df -h`
shows less than the box promised, and nothing is missing.

Kubernetes uses both. `memory: 512Mi` means 512 MiB; `memory: 512M` means
512 MB, about 5% less. Mix them up and a pod can be killed for using too
much memory sooner than you planned.

## Why types exist

Look at this byte again:

```text
0 1 0 0 0 0 0 1
```

Is it the number 65, or the letter `A`? **Both.** In the text code computers
use, `A` is stored as 65. The bits alone don't say which you meant.

It gets worse with more bytes. The two bytes `0x1F 0x90` are the number 8080.
The four characters `"8080"` are four different bytes, one per digit:
`0x38 0x30 0x38 0x30`. (`0x` means the number is written in hexadecimal, a
short way to write bytes.)

A *type* settles it. It says how many bytes a value takes, what they mean,
and which operations make sense. You can multiply two numbers. You can make
text uppercase. Uppercasing a number, or multiplying two names, is nonsense,
and Go refuses.

The types you'll use in *Basics*:

| Type | Holds | Examples |
|---|---|---|
| `int` | Whole numbers, positive or negative | `3`, `8080`, `-1` |
| `float64` | Numbers with a fractional part | `0.5`, `99.9`, `2.0` |
| `string` | Text | `"api"`, `"8080"`, `""` |
| `bool` | Yes or no | `true`, `false` |

## Literals and names: "8080", 8080 and port

This is the idea that trips up almost everyone at first, so take it slowly.

### Three ways to write a value

```go
"8080" // a string literal: the four characters 8, 0, 8, 0
8080   // an int literal: the number eight thousand and eighty
port   // a name: whatever value is stored under port right now
```

A *literal* is a value written straight into your code. A *name* (Go's
word is *identifier*) is a label that points at a value stored somewhere
else.

Picture a labelled box:

```text
port := 8080

   the name (label)          the value (what's inside)
  ┌────────────────┐        ┌────────────────┐
  │      port      │ ─────▶ │      8080      │   type: int
  └────────────────┘        └────────────────┘
```

When Go meets the name `port`, it opens the box and uses what's inside.

### Quotes mean "exactly these characters"

Double quotes switch Go into text mode. Everything between them is taken
literally, letter for letter. Outside quotes, Go reads words as names and
keywords, and looks each name up.

| You write | Go reads it as | Type |
|---|---|---|
| `"port"` | The four letters p, o, r, t | `string` |
| `port` | The value stored under the name `port` | whatever `port` holds |
| `"8080"` | The four characters 8, 0, 8, 0 | `string` |
| `8080` | The number 8080 | `int` |

Here it is running:

```go
package main

import "fmt"

func main() {
	port := 8080

	fmt.Println("port")       // the four letters p, o, r, t
	fmt.Println(port)         // the value stored under the name port
	fmt.Println("8080" + "1") // text joined to text
	fmt.Println(8080 + 1)     // a number plus a number
	fmt.Println(port + 1)     // the value of port, plus one
}
```

It prints:

```text
port
8080
80801
8081
8081
```

Step by step:

1. `port := 8080` makes a box labelled `port` and puts the number 8080 in.
2. `"port"` is in quotes, so it's just text. Go prints the word.
3. `port` has no quotes, so Go opens the box and prints `8080`.
4. `"8080" + "1"`: both are text. `+` on strings glues them: `80801`.
5. `8080 + 1`: both are numbers. `+` on numbers adds: `8081`.
6. `port + 1`: the box holds a number, so this adds too.

**The same `+` does different things depending on the types.** That's types
at work.

### Spaces and punctuation inside quotes count

Inside quotes, a space is a character like any other. So is a dash, a dot or
a colon. Go adds nothing and removes nothing.

Say `namespace` holds `"prod"`:

```go
"kubectl get pods -n " + namespace // "kubectl get pods -n prod"   ✓
"kubectl get pods -n namespace"    // "kubectl get pods -n namespace"
"kubectl get pods -n" + namespace  // "kubectl get pods -nprod"
```

The second line has `namespace` inside the quotes, so it's the word, not the
value. The third line forgot the space before the closing quote.

When you can't see what's wrong, print the value with `fmt.Printf("%q\n",
s)`. The `%q` verb wraps it in quotes, so a missing or extra space at either
end shows up.

### Which one do I need?

Ask this about each piece of the value you're building:

| Question | If yes, write | Example |
|---|---|---|
| Is it exact text that's the same every time? | A string literal | `"kubectl get pods -n "` |
| Does it change with the input or a setting? | A name, no quotes | `namespace` |
| Is it a number you'll calculate with? | A number, no quotes | `8080`, `replicas` |
| Is it digits that are really a label? | A string | `"0042"` (an ID), `"v1.2.3"` |

### The compiler catches type mix-ups

Because every value has a type, Go notices when you mix them:

```go
package main

import "fmt"

func main() {
	var port int
	port = "8080"
	fmt.Println("port: " + port)
}
```

```text
./main.go:7:9: cannot use "8080" (untyped string constant) as int value in assignment
./main.go:8:14: invalid operation: "port: " + port (mismatched types untyped string and int)
```

Line 7: `port` is an `int` box, and `"8080"` is text. Fix: drop the quotes.
Line 8: you can't glue text to a number. Fix: turn the number into text first
(see *Static typing and conversion* below) or use `fmt.Sprintf("port: %d", port)`.

The *Reading compiler errors* exercise practises exactly these messages.

## Whole numbers: int

### Integer division throws away the remainder

An `int` holds whole numbers only, so dividing two ints gives a whole
number. The fractional part is dropped, not rounded: `7 / 2` is `3`. Use
`%` for the remainder: `7 % 2` is `1`.

That drop bites when you work out percentages:

```go
package main

import "fmt"

func main() {
	used := 90   // GiB
	total := 200 // GiB

	fmt.Println(used / total * 100)
	fmt.Println(used * 100 / total)
	fmt.Println(float64(used) / float64(total) * 100)
}
```

It prints:

```text
0
45
45
```

Step by step: `90 / 200` is 0.45, which as an int is `0`, and `0 * 100` is
still `0`. Multiplying first gives `9000 / 200`, which is exactly `45`. The
last line converts both to `float64`, so nothing is dropped.

Picture a disk alert that checks "used / total * 100 > 90". It reads 0% until
the disk is completely full.

### Overflow: when the box is full

An `int` is 64 bits, so it has a largest value: 9,223,372,036,854,775,807.
Go does not grow the box. Add one more and the number *overflows*: it wraps
around to the most negative value.

```go
package main

import "fmt"

func main() {
	big := 9223372036854775807 // the largest int
	big = big + 1
	fmt.Println(big)
}
```

It prints `-9223372036854775808`. No error, no warning.

With 64 bits you'll rarely get there. Smaller boxes fill up fast, though.
Many systems store time as seconds since 1970 in 32 bits, and that count
runs out in January 2038. A counter in a 16-bit box wraps after 65,535.

## Fractions: float64

### 0.1 + 0.2 is not quite 0.3

A `float64` stores a number with a fractional part in 64 bits. It's very
close to exact, but not exact:

```go
package main

import "fmt"

func main() {
	a := 0.1
	b := 0.2
	fmt.Println(a + b)
}
```

It prints `0.30000000000000004`.

Why? A float is stored in binary. In binary, 0.1 never ends (like 1/3 in
decimal: 0.3333...), so Go keeps the nearest value it can fit. The tiny
errors in 0.1 and 0.2 add up. This isn't a Go bug; almost every language
does the same.

(Writing `0.1 + 0.2` directly, without variables, prints `0.3`. Go works out
sums of literals exactly while compiling, before turning the result into a
float.)

### When to use which

| Use `int` for | Use `float64` for |
|---|---|
| Counts: replicas, pods, retries | Measurements: CPU cores (`2.5`), load |
| Bytes, MiB, ports | Averages and ratios |
| Money, stored in cents | Percentages you'll show to a person |

Rule of thumb: if it's counted, use `int`. If it's measured, use `float64`,
and round only when you display it, as in `fmt.Sprintf("%.1f", pct)`.

Durations are a nice example. Go's `time.Duration` is really a whole number
of nanoseconds in 64 bits: exact, and good for about 290 years. You'll use
it in *Time and durations*.

## Text: string

A *string* is a row of bytes holding text. `len` counts the bytes, not the
letters:

```go
package main

import "fmt"

func main() {
	fmt.Println(len("api"), len("café"), len("🚀"))
}
```

It prints `3 5 4`.

Go stores text as *UTF-8*, a standard code that gives every character in
every language a number. Plain English letters, digits and punctuation take
one byte each. `é` takes two bytes. An emoji takes four.

For hostnames, log levels and kubectl commands, one byte per character is
what you'll see. Once you handle user names or log messages from around the
world, it isn't. *Bytes, runes and UTF-8* in the *Collections* module covers
that properly.

## Static typing and conversion

### Types are fixed at compile time

Go is *statically typed*: every variable gets a type when it's created, and
that type never changes. `port := 8080` makes an `int` forever. Go works
out the type from the value, and the compiler checks every use of `port`
against it.

Python and bash check types (or don't) while running. That's why a Go typo
fails on your laptop, and a bash one fails halfway through a deploy (see the
*How a program runs* guide).

### Converting is your job

Go never quietly turns one type into another. You ask, by writing the type
like a function:

```go
replicas := 3
f := float64(replicas) // 3.0, a float64

cpu := 2.9
n := int(cpu) // 2: the fraction is cut off, not rounded
```

Turning a number into text is different. `strconv.Itoa(8080)` gives the
string `"8080"`. Don't write `string(8080)`: it treats the number as a
character code, as in "65 is `A`", and gives you a strange symbol. `go vet`
flags it, and so does `go test`.

Why so strict? Silent conversions hide bugs. If Go quietly turned `90 / 200`
into a float for you here and not there, you'd never know which you got.
Making you write `float64(...)` puts the decision where you can see it.

## Zero values

What's in a box you created without a value? In Go, never garbage. Every
type has a *zero value*:

```go
package main

import "fmt"

func main() {
	var port int
	var cpu float64
	var ready bool
	var host string

	fmt.Println(port, cpu, ready)
	fmt.Printf("%q\n", host)
}
```

It prints:

```text
0 0 false
""
```

| Type | Zero value |
|---|---|
| `int` | `0` |
| `float64` | `0` |
| `bool` | `false` |
| `string` | `""` (empty) |

This is safe, but it can hide a missing setting. A config with no port set
gives you `0`, and port 0 means "any free port" to the OS, so your server
starts somewhere nobody expects. You'll handle this in *Zero values: default
the config*.

## Common mistakes

- **Quoting a name.** `"namespace"` is the word, not the value. Names never
  go in quotes.
- **Forgetting the space inside the quotes.** `"-n" + ns` gives `-nprod`.
  Print with `%q` to see it.
- **Dividing ints for a percentage.** `used / total * 100` is 0 until the
  disk is full. Multiply first, or convert to `float64`.
- **`string(n)` to make digits.** It gives a character, not `"8080"`. Use
  `strconv.Itoa`.
- **Comparing floats exactly.** `0.1 + 0.2 == 0.3` is false for float
  variables. Compare with a small allowed difference, or use ints.
- **Mixing MiB and MB.** `512M` in a Kubernetes manifest is not `512Mi`.
- **A leading zero.** In Go, `0644` is octal (base 8), the way file
  permissions are written: it equals 420. So `0042` is 34, not 42.

## Check yourself

1. What does `"8080" + "1"` give, and what does `8080 + 1` give? Why are
   they different?
2. `namespace` holds `"staging"`. What does
   `"kubectl get pods -n namespace"` give?
3. A disk has 30 GiB used out of 120. What does `used / total * 100` give
   with ints, and how do you get 25?
4. How many different values fit in 16 bits, and where have you seen that
   limit?
5. You declare `var timeout int` and never set it. What value does it have?
   Why might that be a problem?

## Answers

1. `"80801"` and `8081`. Both sides of the first are strings, so `+` glues
   the text. Both sides of the second are ints, so `+` adds. The type
   decides what the operator does.
2. `"kubectl get pods -n namespace"`, the word itself. `namespace` is inside
   the quotes, so it's text. You want `"kubectl get pods -n " + namespace`.
3. `0`: `30 / 120` is 0.25, which as an int is 0. Write
   `used * 100 / total` (3000 / 120 = 25), or convert both to `float64`
   first.
4. 2 to the power 16, which is 65,536: the values 0 to 65,535. That's the
   range of port numbers, which travel in 16 bits.
5. `0`, the zero value for `int`. A timeout of 0 might mean "give up
   instantly" or "never time out", depending on the code that reads it, and
   neither is what you meant. Missing settings deserve a real default.

## Key terms

| Term | Meaning |
|---|---|
| Bit | One 0-or-1 switch: the smallest piece of data. |
| Byte | Eight bits; holds a value from 0 to 255. |
| Binary | Writing numbers in base 2, using only 0 and 1. |
| MiB, GiB | 1,048,576 and 1,073,741,824 bytes (powers of 2). |
| Type | What a value's bytes mean and what you can do with it. |
| Literal | A value written straight into code: `"api"`, `8080`, `true`. |
| Name (identifier) | A label that points at a stored value: `port`. |
| Integer division | Dividing ints: the fractional part is dropped. |
| Overflow | A number going past its type's limit and wrapping around. |
| Float | A number with a fractional part, stored approximately. |
| String | A row of bytes holding text. |
| UTF-8 | The standard way to store any character as 1 to 4 bytes. |
| Statically typed | Every variable's type is fixed and checked at compile time. |
| Conversion | Turning a value into another type on purpose: `float64(n)`. |
| Zero value | The value a variable has before you set one: `0`, `""`, `false`. |

## Go deeper

- [Wikipedia: Binary number](https://en.wikipedia.org/wiki/Binary_number)
  — more on counting in base 2, with worked examples.
- [Floating Point Math](https://0.30000000000000004.com/) — the same
  0.1 + 0.2 result in dozens of languages, with a short explanation of why.
- [Strings, bytes, runes and characters in Go](https://go.dev/blog/strings)
  — the Go blog on how strings really work; read it before *Bytes, runes and
  UTF-8*.
- [Constants](https://go.dev/blog/constants) — why `0.1 + 0.2` written as
  literals is exact, and how untyped constants work.
