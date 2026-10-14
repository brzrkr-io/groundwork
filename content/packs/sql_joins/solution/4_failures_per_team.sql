-- Task 4: Count failed deploys per team
--
-- Return each team with at least one failed deploy: its name (team)
-- and its number of failed deploys (failures), A to Z by team.

SELECT t.name AS team, COUNT(*) AS failures
FROM deploys AS d
JOIN services AS s ON s.id = d.service_id
JOIN teams AS t ON t.id = s.team_id
WHERE d.status = 'failed'
GROUP BY t.name
ORDER BY t.name;
