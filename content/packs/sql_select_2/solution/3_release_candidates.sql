-- Task 3: Fix the release candidate search
--
-- Show the service and version of every release candidate: a
-- version with -rc in it. This returns no rows at all.

SELECT service, version
FROM deploys
WHERE version LIKE '%-rc%';
