-- Rollback: Drop patient_profiles table
DROP INDEX IF EXISTS idx_patient_profiles_phone_number;
DROP INDEX IF EXISTS idx_patient_profiles_nik;
DROP TABLE IF EXISTS patient_profiles;
