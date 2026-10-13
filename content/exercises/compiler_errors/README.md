# Reading compiler errors

The Go compiler checks your code before anything runs. When it finds a
problem it stops, and tells you exactly where it is and what's wrong. Once you
can read these messages, most of them take seconds to fix. Here you practise
on the four you'll meet most.

**New here:** reading `file:line:column` messages; finding the real cause; a
method that works for any error.

## Learn

### The shape of an error

```text
./exercise.go:12:22: undefined: hostname
└────┬──────┘ └┬┘└┬┘ └──────────┬─────────┘
   file      line column     what's wrong
```

Go to that **line** first. The **column** counts characters from the left and
points at the exact word. Gopherlings shows errors in this form when your code
doesn't compile.

### A method for any compile error

1. **Fix the first error first.** One mistake can cause several messages, so
   later ones may vanish once the first is fixed. Save after each fix and read
   the list again.
2. **Go to the line.** Read the line out loud, word by word.
3. **Split the message into "what Go found" and "what Go expected".** Most
   messages have that shape: *cannot use **this** as **that***, *mismatched
   types **A** and **B***.
4. **Follow names back.** If the line looks fine, find where each name on it
   was created: a parameter in the function's `( )`, a constant, or a function.
   The mistake is often *there*, not on the line Go reported.

> **Why is it written like this?** Go checks the whole package before running
> a single line, so it reports every problem it can find at once, ordered by
> line. The compiler can't know what you *meant*, only that two things don't
> fit. That's why the message names both sides.

### The four messages in this exercise

| Message | In plain words |
|---|---|
| `undefined: x` | Go can't find anything called `x` here. It's usually a typo, or a name that belongs to a different function. |
| `mismatched types string and func() string` | You're combining text with a *function*. A function's name alone is the function itself, not what it returns. |
| `cannot use "8080" (untyped string constant) as int value` | You gave text (it has quotes) where a number is needed. |
| `declared and not used: x` | You made a variable and never used it. Go treats that as a mistake, because it usually is one. |

> **Why is it written like this?** "Declared and not used" is an error in Go,
> not a warning. An unused variable is almost always a bug: you meant to use
> it and used something else. Go makes you deal with it now.

## Your task

When you start, Gopherlings shows four errors, one per task, in line order.
Fix them one at a time. After each save, the list gets shorter. When the code
compiles, the tests run and check that each function returns the right thing.

### 1. Fix the undefined name

**Goal:** make `URL` compile and return `https://` followed by the host.

```text
./exercise.go:12:22: undefined: hostname
```

```go
URL("api.internal") // => "https://api.internal"
URL("grafana.ops")  // => "https://grafana.ops"
```

**Think it through:**
- Which names can `URL` use? Look inside the `( )` on its first line.
- Compare that name with the one on line 12, letter by letter.

### 2. Fix the mismatched types

**Goal:** make `Tag` compile and return the service, a colon, and the release
from `Version`.

```text
./exercise.go:18:9: invalid operation: service + ":" + Version (mismatched types string and func() string)
```

```go
Tag("api")    // => "api:v2"
Tag("worker") // => "worker:v2"
```

**Think it through:**
- `Version` is declared at the top of the file. Is it a constant or a
  function? Look at the word it starts with.
- Do you want the function itself, or the text it returns? What do you add to
  a name to run a function?

### 3. Fix the text-instead-of-number

**Goal:** make `Port` compile and return the number `8080`.

```text
./exercise.go:24:9: cannot use "8080" (untyped string constant) as int value in return statement
```

```go
Port() // => 8080
```

**Think it through:**
- What does `Port` promise to return? Read the end of its first line.
- What makes `"8080"` text instead of a number?

### 4. Fix the unused variable

**Goal:** make `Greeting` compile and return `Hello, ` followed by the name.

```text
./exercise.go:30:2: declared and not used: full
```

```go
Greeting("Sam")   // => "Hello, Sam"
Greeting("Priya") // => "Hello, Priya"
```

**Think it through:**
- `full` already holds the right text. Where does the function use it?
- What does the `return` line give back instead?

## Learn more

- [Go by Example: Variables](https://gobyexample.com/variables)
- [Go FAQ: Why does Go report unused variables?](https://go.dev/doc/faq#unused_variables_and_imports)
- [A Tour of Go: Functions](https://go.dev/tour/basics/4)
