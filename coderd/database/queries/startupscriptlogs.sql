-- name: GetWorkspaceAgentStartupLogsByID :one
SELECT
	*
FROM
	startup_script_logs
WHERE
	agent_id = $1;

-- name: InsertOrUpdateWorkspaceAgentStartupLogsByID :exec
INSERT INTO
	startup_script_logs (agent_id, output)
VALUES ($1, $2)
ON CONFLICT (agent_id) DO UPDATE
	SET
		output = $2
	WHERE
		agent_id = $1;
