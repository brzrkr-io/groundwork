-- Task 4: Rewrite the query so it can use the index
--
-- Return the id and service of every deploy that started on
-- 2026-03-04, earliest first. The answer is right, but wrapping
-- started_at in date() hides it from the deploys_by_start index,
-- so SQLite reads every row.
--
-- TODO(4): Rewrite the WHERE so the deploys_by_start index can be used.

SELECT id, service
FROM deploys
WHERE date(started_at) = '2026-03-04'
ORDER BY started_at;
