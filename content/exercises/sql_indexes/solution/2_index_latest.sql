-- Task 2: Index the latest deploys of a service
--
-- Create an index named deploys_by_service_time that serves this
-- query without reading every row and without a separate sort:
--
--     SELECT id, version FROM deploys
--     WHERE service = 'api'
--     ORDER BY started_at DESC
--     LIMIT 5;

CREATE INDEX deploys_by_service_time ON deploys (service, started_at);
