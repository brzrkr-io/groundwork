-- Task 2: Find the failed deploys
--
-- Show the service and version of every deploy whose status is
-- 'failed'.

SELECT service, version
FROM deploys
WHERE status = 'failed';
