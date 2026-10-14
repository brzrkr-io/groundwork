-- Task 2: Fix the non-production filter
--
-- Return the id, service and environment of every deploy that is not
-- known to be a production deploy, including those whose environment
-- was never recorded, by id. This misses the unrecorded ones.
--
-- TODO(2): Fix the filter so deploys with no environment are included.

SELECT id, service, environment
FROM deploys
WHERE environment != 'production'
ORDER BY id;
