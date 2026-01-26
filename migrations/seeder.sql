-- Seeder Script for RS Azra Hospital Booking System
-- Run this after migrations to seed initial data

-- Insert roles
INSERT INTO roles (role_name, description) VALUES
    ('admin', 'System administrator with full access'),
    ('doctor', 'Medical professional who provides consultations'),
    ('patient', 'Patient who books appointments')
ON CONFLICT (role_name) DO NOTHING;

-- Insert admin user
-- Password: admin123 (bcrypt hashed with cost 10)
-- Note: In production, change this password immediately
INSERT INTO users (id, role_id, email, password, full_name, is_active)
SELECT 
    gen_random_uuid(),
    r.id,
    'admin@rsazra.co.id',
    '$2a$10$N9qo8uLOickgx2ZMRZoMye1pN3n9H1JgL/7X6YBPg8GAK.NDAYoAK',
    'Super Admin RS Azra',
    true
FROM roles r
WHERE r.role_name = 'admin'
ON CONFLICT (email) DO NOTHING;

-- Verification query (optional, comment out in production)
-- SELECT u.id, u.email, u.full_name, r.role_name 
-- FROM users u 
-- JOIN roles r ON u.role_id = r.id;
