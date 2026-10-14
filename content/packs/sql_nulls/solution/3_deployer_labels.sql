-- Task 3: Fix the deployer labels
--
-- Return every deploy's id and a label: 'deployed by ' followed by
-- deployed_by, or by 'automation' when deployed_by is empty, by id.
-- This gives NULL labels for the pipeline's deploys.

SELECT id, 'deployed by ' || COALESCE(deployed_by, 'automation') AS label
FROM deploys
ORDER BY id;
