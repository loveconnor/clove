ALTER TABLE IF EXISTS sessions DROP CONSTRAINT IF EXISTS sessions_user_id_fkey;
ALTER TABLE IF EXISTS organizations DROP CONSTRAINT IF EXISTS organizations_owner_id_fkey;
ALTER TABLE IF EXISTS organization_members DROP CONSTRAINT IF EXISTS organization_members_organization_id_fkey;
ALTER TABLE IF EXISTS organization_members DROP CONSTRAINT IF EXISTS organization_members_user_id_fkey;
ALTER TABLE IF EXISTS ssh_keys DROP CONSTRAINT IF EXISTS ssh_keys_user_id_fkey;
ALTER TABLE IF EXISTS audit_logs DROP CONSTRAINT IF EXISTS audit_logs_actor_user_id_fkey;
ALTER TABLE IF EXISTS audit_logs DROP CONSTRAINT IF EXISTS audit_logs_organization_id_fkey;
ALTER TABLE IF EXISTS audit_logs DROP CONSTRAINT IF EXISTS audit_logs_repository_id_fkey;

ALTER TABLE users ALTER COLUMN id TYPE TEXT USING id::TEXT;
ALTER TABLE sessions ALTER COLUMN id TYPE TEXT USING id::TEXT;
ALTER TABLE sessions ALTER COLUMN user_id TYPE TEXT USING user_id::TEXT;
ALTER TABLE organizations ALTER COLUMN id TYPE TEXT USING id::TEXT;
ALTER TABLE organizations ALTER COLUMN owner_id TYPE TEXT USING owner_id::TEXT;
ALTER TABLE organization_members ALTER COLUMN organization_id TYPE TEXT USING organization_id::TEXT;
ALTER TABLE organization_members ALTER COLUMN user_id TYPE TEXT USING user_id::TEXT;
ALTER TABLE repositories ALTER COLUMN id TYPE TEXT USING id::TEXT;
ALTER TABLE repositories ALTER COLUMN owner_id TYPE TEXT USING owner_id::TEXT;
ALTER TABLE ssh_keys ALTER COLUMN id TYPE TEXT USING id::TEXT;
ALTER TABLE ssh_keys ALTER COLUMN user_id TYPE TEXT USING user_id::TEXT;
ALTER TABLE audit_logs ALTER COLUMN id TYPE TEXT USING id::TEXT;
ALTER TABLE audit_logs ALTER COLUMN actor_user_id TYPE TEXT USING actor_user_id::TEXT;
ALTER TABLE audit_logs ALTER COLUMN organization_id TYPE TEXT USING organization_id::TEXT;
ALTER TABLE audit_logs ALTER COLUMN repository_id TYPE TEXT USING repository_id::TEXT;
ALTER TABLE audit_logs ALTER COLUMN target_id TYPE TEXT USING target_id::TEXT;

ALTER TABLE sessions
	ADD CONSTRAINT sessions_user_id_fkey
	FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE organizations
	ADD CONSTRAINT organizations_owner_id_fkey
	FOREIGN KEY (owner_id) REFERENCES users(id);

ALTER TABLE organization_members
	ADD CONSTRAINT organization_members_organization_id_fkey
	FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE;

ALTER TABLE organization_members
	ADD CONSTRAINT organization_members_user_id_fkey
	FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE ssh_keys
	ADD CONSTRAINT ssh_keys_user_id_fkey
	FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE audit_logs
	ADD CONSTRAINT audit_logs_actor_user_id_fkey
	FOREIGN KEY (actor_user_id) REFERENCES users(id) ON DELETE SET NULL;

ALTER TABLE audit_logs
	ADD CONSTRAINT audit_logs_organization_id_fkey
	FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE SET NULL;

ALTER TABLE audit_logs
	ADD CONSTRAINT audit_logs_repository_id_fkey
	FOREIGN KEY (repository_id) REFERENCES repositories(id) ON DELETE SET NULL;
