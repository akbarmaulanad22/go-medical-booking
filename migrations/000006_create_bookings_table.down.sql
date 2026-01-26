-- Rollback: Drop bookings table and enum type
DROP INDEX IF EXISTS idx_bookings_patient_status;
DROP INDEX IF EXISTS idx_bookings_status;
DROP INDEX IF EXISTS idx_bookings_schedule_id;
DROP INDEX IF EXISTS idx_bookings_patient_id;
DROP INDEX IF EXISTS idx_bookings_booking_code;
DROP TABLE IF EXISTS bookings;
DROP TYPE IF EXISTS booking_status;
