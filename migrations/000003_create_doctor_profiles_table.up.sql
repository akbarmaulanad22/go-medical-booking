-- Migration: Create doctor_profiles table
-- Description: Stores detailed profile information for doctors

CREATE TABLE IF NOT EXISTS doctor_profiles (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    str_number VARCHAR(50) NOT NULL UNIQUE,
    specialization VARCHAR(100) NOT NULL,
    biography TEXT
);

-- Index for STR number lookups
CREATE INDEX IF NOT EXISTS idx_doctor_profiles_str_number ON doctor_profiles(str_number);

-- Index for specialization searches
CREATE INDEX IF NOT EXISTS idx_doctor_profiles_specialization ON doctor_profiles(specialization);

COMMENT ON TABLE doctor_profiles IS 'Doctor-specific profile data linked to users table';
COMMENT ON COLUMN doctor_profiles.str_number IS 'Surat Tanda Registrasi - Doctor registration number';
