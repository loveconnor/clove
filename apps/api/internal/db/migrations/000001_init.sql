CREATE TABLE IF NOT EXISTS users (
	id UUID PRIMARY KEY,
	username TEXT NOT NULL UNIQUE,
	email TEXT NOT NULL UNIQUE,
	display_name TEXT,
	avatar_url TEXT,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS sessions (
	id UUID PRIMARY KEY,
	user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	token_hash TEXT NOT NULL UNIQUE,
	user_agent TEXT,
	ip_address INET,
	expires_at TIMESTAMPTZ NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	last_used_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS sessions_user_id_idx ON sessions(user_id);
CREATE INDEX IF NOT EXISTS sessions_expires_at_idx ON sessions(expires_at);

CREATE TABLE IF NOT EXISTS organizations (
	id UUID PRIMARY KEY,
	name TEXT NOT NULL UNIQUE,
	display_name TEXT,
	owner_id UUID NOT NULL REFERENCES users(id),
	created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS organization_members (
	organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
	user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	role TEXT NOT NULL CHECK (role IN ('owner', 'admin', 'member')),
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	PRIMARY KEY (organization_id, user_id)
);

CREATE INDEX IF NOT EXISTS organization_members_user_id_idx ON organization_members(user_id);

CREATE TABLE IF NOT EXISTS repositories (
	id UUID PRIMARY KEY,
	owner_type TEXT NOT NULL CHECK (owner_type IN ('user', 'organization')),
	owner_id UUID NOT NULL,
	name TEXT NOT NULL,
	description TEXT,
	visibility TEXT NOT NULL CHECK (visibility IN ('public', 'private', 'internal')),
	default_branch TEXT NOT NULL DEFAULT 'main',
	git_path TEXT NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	UNIQUE(owner_id, name)
);

CREATE INDEX IF NOT EXISTS repositories_owner_idx ON repositories(owner_type, owner_id);
CREATE INDEX IF NOT EXISTS repositories_visibility_idx ON repositories(visibility);

CREATE TABLE IF NOT EXISTS ssh_keys (
	id UUID PRIMARY KEY,
	user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	name TEXT NOT NULL,
	public_key TEXT NOT NULL UNIQUE,
	fingerprint TEXT NOT NULL UNIQUE,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	last_used_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS ssh_keys_user_id_idx ON ssh_keys(user_id);

CREATE TABLE IF NOT EXISTS audit_logs (
	id UUID PRIMARY KEY,
	actor_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
	organization_id UUID REFERENCES organizations(id) ON DELETE SET NULL,
	repository_id UUID REFERENCES repositories(id) ON DELETE SET NULL,
	action TEXT NOT NULL,
	target_type TEXT,
	target_id UUID,
	metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS audit_logs_actor_user_id_idx ON audit_logs(actor_user_id);
CREATE INDEX IF NOT EXISTS audit_logs_organization_id_idx ON audit_logs(organization_id);
CREATE INDEX IF NOT EXISTS audit_logs_repository_id_idx ON audit_logs(repository_id);
CREATE INDEX IF NOT EXISTS audit_logs_created_at_idx ON audit_logs(created_at);
