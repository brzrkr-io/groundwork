-- Task 3: Fix the deployer labels
--
-- Return every deploy's id and a label: 'deployed by ' followed by
-- deployed_by, or by 'automation' when deployed_by is NULL, by id.

SELECT id, 'deployed by ' || COALESCE(deployed_by, 'automation') AS label
FROM deploys
ORDER BY id;
