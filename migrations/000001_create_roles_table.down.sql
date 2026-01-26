-- Rollback: Drop roles table
DROP INDEX IF EXISTS idx_roles_role_name;
DROP TABLE IF EXISTS roles;
