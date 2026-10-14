-- Task 3: Find slow production deploys
--
-- Show the id, service and duration_s of every production deploy
-- that took longer than 300 seconds.

SELECT id, service, duration_s
FROM deploys
WHERE environment = 'production'
  AND duration_s > 300;
