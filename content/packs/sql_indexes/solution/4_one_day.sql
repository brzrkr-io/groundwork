-- Task 4: Rewrite the query so it can use the index
--
-- Return the id and service of every deploy that started on
-- 2026-03-04, earliest first. Compare the bare started_at column with
-- a range so the deploys_by_start index can be searched.

SELECT id, service
FROM deploys
WHERE started_at >= '2026-03-04'
  AND started_at < '2026-03-05'
ORDER BY started_at;
