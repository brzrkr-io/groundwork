-- Task 3: Find the Slack channel to alert
--
-- Return the id, service name (service), status and the owning team's
-- slack_channel of every deploy that did not succeed, by deploy id.

SELECT d.id, s.name AS service, d.status, t.slack_channel
FROM deploys AS d
JOIN services AS s ON s.id = d.service_id
JOIN teams AS t ON t.id = s.team_id
WHERE d.status != 'succeeded'
ORDER BY d.id;
