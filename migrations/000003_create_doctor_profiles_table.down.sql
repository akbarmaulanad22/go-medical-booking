-- Rollback: Drop doctor_profiles table
DROP INDEX IF EXISTS idx_doctor_profiles_specialization;
DROP INDEX IF EXISTS idx_doctor_profiles_str_number;
DROP TABLE IF EXISTS doctor_profiles;
