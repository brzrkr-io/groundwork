-- Task 1: Index the service column
--
-- Create an index named deploys_by_service on the service column of
-- deploys, so WHERE service = '...' stops reading every row.

CREATE INDEX deploys_by_service ON deploys (service);
