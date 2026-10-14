-- Task 3: Stop duplicate deploy records
--
-- Create a unique index named deploys_one_per_release so the same
-- version of a service can't be recorded twice for one environment.

CREATE UNIQUE INDEX deploys_one_per_release ON deploys (service, version, environment);
