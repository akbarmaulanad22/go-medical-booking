-- Rollback: Remove partial unique constraint for duplicate booking prevention

DROP INDEX IF EXISTS idx_bookings_patient_schedule_active;
