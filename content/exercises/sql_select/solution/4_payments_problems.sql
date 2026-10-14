-- Task 4: Find the payments team's problem deploys
--
-- Show the id, service and status of every deploy of checkout or
-- billing that did not succeed.

SELECT id, service, status
FROM deploys
WHERE service IN ('checkout', 'billing')
  AND status != 'succeeded';
