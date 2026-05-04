CREATE TABLE IF NOT EXISTS git_refs (
	repo_id TEXT NOT NULL REFERENCES repositories(id) ON DELETE CASCADE,
	ref_name TEXT NOT NULL,
	old_sha TEXT NOT NULL,
	new_sha TEXT NOT NULL,
	pusher_id TEXT REFERENCES users(id) ON DELETE SET NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	PRIMARY KEY (repo_id, ref_name)
);

CREATE INDEX IF NOT EXISTS git_refs_repo_id_idx ON git_refs(repo_id);
CREATE INDEX IF NOT EXISTS git_refs_pusher_id_idx ON git_refs(pusher_id);

CREATE TABLE IF NOT EXISTS push_events (
	id TEXT PRIMARY KEY,
	repo_id TEXT NOT NULL REFERENCES repositories(id) ON DELETE CASCADE,
	ref_name TEXT NOT NULL,
	old_sha TEXT NOT NULL,
	new_sha TEXT NOT NULL,
	pusher_id TEXT REFERENCES users(id) ON DELETE SET NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS push_events_repo_id_idx ON push_events(repo_id);
CREATE INDEX IF NOT EXISTS push_events_pusher_id_idx ON push_events(pusher_id);
CREATE INDEX IF NOT EXISTS push_events_created_at_idx ON push_events(created_at);
