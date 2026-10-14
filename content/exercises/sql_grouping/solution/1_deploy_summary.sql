-- Task 1: Summarize every deploy
--
-- Return one row: how many deploys there are (deploys), the longest
-- duration_s (slowest_s), and the earliest started_at (first_started).

SELECT
    COUNT(*) AS deploys,
    MAX(duration_s) AS slowest_s,
    MIN(started_at) AS first_started
FROM deploys;
