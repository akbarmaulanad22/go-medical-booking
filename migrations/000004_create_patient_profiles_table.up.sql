-- Migration: Create patient_profiles table
-- Description: Stores detailed profile information for patients

CREATE TABLE IF NOT EXISTS patient_profiles (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    nik CHAR(16) NOT NULL UNIQUE,
    phone_number VARCHAR(20),
    date_of_birth DATE NOT NULL,
    gender CHAR(1) NOT NULL CHECK (gender IN ('M', 'F')),
    address TEXT
);

-- Index for NIK lookups
CREATE UNIQUE INDEX IF NOT EXISTS idx_patient_profiles_nik ON patient_profiles(nik);

-- Index for phone number searches
CREATE INDEX IF NOT EXISTS idx_patient_profiles_phone_number ON patient_profiles(phone_number);

COMMENT ON TABLE patient_profiles IS 'Patient-specific profile data linked to users table';
COMMENT ON COLUMN patient_profiles.nik IS 'Nomor Induk Kependudukan - Indonesian ID number (16 digits)';
COMMENT ON COLUMN patient_profiles.gender IS 'M = Male, F = Female';
