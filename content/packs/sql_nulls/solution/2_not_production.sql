-- Task 2: Fix the non-production filter
--
-- Return the id, service and environment of every deploy that is not
-- known to be a production deploy, including those whose environment
-- was never recorded, by id. This misses the unrecorded ones.

SELECT id, service, environment
FROM deploys
WHERE environment IS NULL
   OR environment != 'production'
ORDER BY id;
