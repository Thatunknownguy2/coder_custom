CREATE TABLE IF NOT EXISTS startup_script_logs (
    agent_id uuid NOT NULL REFERENCES workspace_agents (id) ON DELETE CASCADE,
    output text NOT NULL
);
