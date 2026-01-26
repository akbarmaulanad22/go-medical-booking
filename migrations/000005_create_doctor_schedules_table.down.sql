-- Rollback: Drop doctor_schedules table
DROP INDEX IF EXISTS idx_doctor_schedules_available;
DROP INDEX IF EXISTS idx_doctor_schedules_schedule_date;
DROP INDEX IF EXISTS idx_doctor_schedules_doctor_id;
DROP TABLE IF EXISTS doctor_schedules;
