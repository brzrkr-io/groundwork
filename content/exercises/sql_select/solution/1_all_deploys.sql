-- Task 1: List every deploy
--
-- Show the service, version and environment of every deploy,
-- in that order.

SELECT service, version, environment
FROM deploys;
