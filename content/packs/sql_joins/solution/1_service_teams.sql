-- Task 1: Show which team owns each service
--
-- Return each service's name (service) and its team's name (team),
-- A to Z by service. Services without a team are left out.

SELECT s.name AS service, t.name AS team
FROM services AS s
JOIN teams AS t ON t.id = s.team_id
ORDER BY s.name;
