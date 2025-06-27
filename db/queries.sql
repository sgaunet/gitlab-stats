-- name: InsertNewStats :one
INSERT INTO stats (total,closed,opened,date_exec)
VALUES(?,?,?,?)
RETURNING id;

-- name: InsertStatsProjects :one
INSERT INTO stats_projects (projectId,statsId)
VALUES(?,?)
RETURNING id;

-- name: InsertStatsGroups :one
INSERT INTO stats_groups (groupId,statsId)
VALUES(?,?)
RETURNING id;

-- name: GetProject :one
SELECT id,project_name FROM projects WHERE id=?;

-- name: GetGroup :one
SELECT id,group_name FROM groups WHERE id=?;

-- name: InsertNewProject :one
INSERT INTO projects (id,project_name)
VALUES(?,?)
RETURNING id;

-- name: InsertNewGroup :one
INSERT INTO groups (id,group_name)
VALUES(?,?)
RETURNING id;

-- name: GetStatsOfProject :many
SELECT id,total,closed,opened,date_exec FROM stats
WHERE id IN (SELECT statsId FROM stats_projects WHERE projectId=?)
ORDER BY date_exec DESC;

-- name: GetStatsOfGroup :many
SELECT id,total,closed,opened,date_exec FROM stats
WHERE id IN (SELECT statsId FROM stats_groups WHERE groupId=?)
ORDER BY date_exec DESC;

-- name: GetStatsByProjectID6Months :many
SELECT id,total,closed,opened,date_exec
FROM (
  SELECT id,total,closed,opened,date_exec,
    ROW_NUMBER() OVER(PARTITION BY strftime('%Y-%m', date_exec) ORDER BY date_exec DESC) as rn
  FROM stats
  order by date_exec
) t
WHERE 
  rn = 2 AND date_exec >= sqlc.arg(begindate) AND date_exec <= sqlc.arg(enddate)
  AND id IN (SELECT statsId FROM stats_projects WHERE projectId=sqlc.arg(projectId))
order by date_exec;

-- name: GetStatsByGroupID6Months :many
SELECT id,total,closed,opened,date_exec
FROM (
  SELECT id,total,closed,opened,date_exec,
    ROW_NUMBER() OVER(PARTITION BY strftime('%Y-%m', date_exec) ORDER BY date_exec DESC) as rn
  FROM stats
  order by date_exec
) t
WHERE 
  rn = 2 AND date_exec >= sqlc.arg(begindate) AND date_exec <= sqlc.arg(enddate)
  AND id IN (SELECT statsId FROM stats_groups WHERE groupId=sqlc.arg(groupId))
order by date_exec;

-- name: GetEnhancedStatsByProjectID :many
SELECT 
  strftime('%Y-%m', date_exec) as period,
  MAX(total) as total_opened,
  MAX(opened) as current_opened,
  MAX(closed) as current_closed,
  LAG(MAX(total), 1, 0) OVER(ORDER BY strftime('%Y-%m', date_exec)) as prev_total,
  LAG(MAX(closed), 1, 0) OVER(ORDER BY strftime('%Y-%m', date_exec)) as prev_closed,
  MAX(date_exec) as date_exec
FROM stats
WHERE 
  date_exec >= sqlc.arg(begindate) AND date_exec <= sqlc.arg(enddate)
  AND id IN (SELECT statsId FROM stats_projects WHERE projectId=sqlc.arg(projectId))
GROUP BY strftime('%Y-%m', date_exec)
ORDER BY period;

-- name: GetEnhancedStatsByGroupID :many
SELECT 
  strftime('%Y-%m', date_exec) as period,
  MAX(total) as total_opened,
  MAX(opened) as current_opened,
  MAX(closed) as current_closed,
  LAG(MAX(total), 1, 0) OVER(ORDER BY strftime('%Y-%m', date_exec)) as prev_total,
  LAG(MAX(closed), 1, 0) OVER(ORDER BY strftime('%Y-%m', date_exec)) as prev_closed,
  MAX(date_exec) as date_exec
FROM stats
WHERE 
  date_exec >= sqlc.arg(begindate) AND date_exec <= sqlc.arg(enddate)
  AND id IN (SELECT statsId FROM stats_groups WHERE groupId=sqlc.arg(groupId))
GROUP BY strftime('%Y-%m', date_exec)
ORDER BY period;
