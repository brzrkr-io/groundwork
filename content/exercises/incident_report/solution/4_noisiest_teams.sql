-- Task 4: Rank the noisiest teams in March
--
-- Return each team with at least 2 incidents opened in March 2026:
-- its name (team) and that number (incidents), most incidents
-- first, then A to Z by team.

SELECT t.name AS team, COUNT(*) AS incidents
FROM incidents AS i
JOIN services AS s ON s.id = i.service_id
JOIN teams AS t ON t.id = s.team_id
WHERE i.opened_at >= '2026-03-01'
  AND i.opened_at < '2026-04-01'
GROUP BY t.id
HAVING COUNT(*) >= 2
ORDER BY incidents DESC, t.name;
