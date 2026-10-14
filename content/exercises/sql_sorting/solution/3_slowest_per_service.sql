-- Task 3: Sort by service, then slowest first
--
-- Show the service and duration_s of every deploy, sorted by service
-- A to Z, and within each service the slowest deploy first.

SELECT service, duration_s
FROM deploys
ORDER BY service, duration_s DESC;
