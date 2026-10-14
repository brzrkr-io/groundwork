-- Task 3: Delete last year's staging deploys
--
-- Delete every staging deploy that started before 2026. Keep all
-- production deploys, however old.

DELETE FROM deploys
WHERE environment = 'staging'
  AND started_at < '2026-01-01';
