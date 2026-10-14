## Task 1: Show which team owns each service

Start from `services`, then bring in `teams`. The `ON` condition matches the
service's `team_id` with the team's `id`. Then name both `name` columns.

### Closer

`FROM services AS s JOIN teams AS t ON t.id = s.team_id`, then select
`s.name AS service, t.name AS team` and `ORDER BY s.name`.

### Answer

```sql
SELECT s.name AS service, t.name AS team
FROM services AS s
JOIN teams AS t ON t.id = s.team_id
ORDER BY s.name;
```

## Task 2: Name the service of each failed deploy

The deploy row only has a `service_id`. Join `services` to turn it into a
name, then filter on the deploy's status.

### Closer

`FROM deploys AS d JOIN services AS s ON s.id = d.service_id`. Select
`d.id`, not `s.id`, and filter with `WHERE d.status = 'failed'`.

### Answer

```sql
SELECT d.id, s.name AS service, d.version
FROM deploys AS d
JOIN services AS s ON s.id = d.service_id
WHERE d.status = 'failed'
ORDER BY d.id;
```

## Task 3: Find the Slack channel to alert

Two hops: from the deploy to its service, then from the service to its team.
Each hop is one `JOIN … ON`.

### Closer

After the `services` join from task 2, add
`JOIN teams AS t ON t.id = s.team_id`, select `t.slack_channel`, and keep
rows with `d.status != 'succeeded'`.

### Answer

```sql
SELECT d.id, s.name AS service, d.status, t.slack_channel
FROM deploys AS d
JOIN services AS s ON s.id = d.service_id
JOIN teams AS t ON t.id = s.team_id
WHERE d.status != 'succeeded'
ORDER BY d.id;
```

## Task 4: Count failed deploys per team

The joins are the same as task 3. After them, keep the failed deploys and
make one group per team.

### Closer

`WHERE d.status = 'failed'`, `GROUP BY t.name`, and select
`t.name AS team, COUNT(*) AS failures`.

### Answer

```sql
SELECT t.name AS team, COUNT(*) AS failures
FROM deploys AS d
JOIN services AS s ON s.id = d.service_id
JOIN teams AS t ON t.id = s.team_id
WHERE d.status = 'failed'
GROUP BY t.name
ORDER BY t.name;
```
