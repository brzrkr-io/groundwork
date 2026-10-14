-- Task 4: Rewrite the query so it can use the index
--
-- Return the id and service of every deploy that started on
-- 2026-03-04, earliest first. The answer is right, but wrapping
-- started_at in date() hides it from the deploys_by_start index,
-- so SQLite reads every row.

SELECT id, service
FROM deploys
WHERE started_at >= '2026-03-04'
  AND started_at < '2026-03-05'
ORDER BY started_at;
