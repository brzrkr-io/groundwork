## Task 1: match verbs to values

The values are `service` (a string), then `from` and `to` (ints). The first
verb must be the string verb.

## Task 2: let %q add the quotes

`%q` puts the message in double quotes for you:
`fmt.Sprintf("level=%s msg=%q duration_ms=%d", level, msg, ms)`.

## Task 3: zero padding

`%05d` pads to 5 digits with zeros. Put `INC-` in front of it in the same
format string.
