-- Task 2: Mark one deploy as rolled back
--
-- Deploy 14 was rolled back. Set its status to rolled_back, and
-- change nothing else.

UPDATE deploys
SET status = 'rolled_back'
WHERE id = 14;
