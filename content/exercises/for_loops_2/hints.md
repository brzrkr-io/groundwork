## Compare got and want

Run the tests and read each failure. For `HostNames`, look at the first number
in the output and at the end of the string. For `BatchCount`, which case is
wrong: the one that divides evenly or the one with leftovers?

## Task 1: two fixes

`for i := range count` gives `i` = 0, 1, 2, ..., but hosts are numbered
from 1, so use `i+1` in the name. Then write the comma *before* each name
except the first (`if i > 0`), instead of after every name.

## Task 2: break versus continue

`break` leaves the loop for good. `continue` only skips this round. Swap one
for the other.

## Task 3: the loop condition

The loop keeps going while `remaining >= batchSize`, so a leftover smaller
than a full batch is never counted. Loop while `remaining > 0` instead.
