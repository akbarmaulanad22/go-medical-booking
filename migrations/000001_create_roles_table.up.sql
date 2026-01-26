-- Migration: Create roles table
-- Description: Stores role information for the shared authentication model

CREATE TABLE IF NOT EXISTS roles (
    id SERIAL PRIMARY KEY,
    role_name VARCHAR(50) NOT NULL UNIQUE,
    description TEXT
);

-- Index for role_name lookups
CREATE INDEX IF NOT EXISTS idx_roles_role_name ON roles(role_name);

COMMENT ON TABLE roles IS 'Stores user roles: admin, doctor, patient';
COMMENT ON COLUMN roles.role_name IS 'Role identifier: admin, doctor, patient';
