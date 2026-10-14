## Task 1: Fix the scale message

Read the vet message. Match each verb to its value in order: `service` is a string, `from` and `to` are ints.

### Closer

The first verb is `%d`, but its value `service` is a string. Change that one verb to the string verb.

### Answer

```go
return fmt.Sprintf("scaling %s from %d to %d replicas", service, from, to)
```

## Task 2: Write a logfmt line

The keys (`level=`, `msg=`, `duration_ms=`) are fixed text in the format string. Which verb puts double quotes around the message for you?

### Closer

Use `%s` for the level, `%q` for the message, `%d` for the duration: `"level=%s msg=%q ..."`. The parameters go after the format string, in order.

### Answer

```go
return fmt.Sprintf("level=%s msg=%q duration_ms=%d", level, msg, durationMs)
```

## Task 3: Format the incident ID

`INC-` is fixed text, then a number padded with zeros. See *Padding with zeros* in the README.

### Closer

`%05d` pads an int to 5 digits with zeros. Put `INC-` in front of it inside the same format string.

### Answer

```go
return fmt.Sprintf("INC-%05d", n)
```
