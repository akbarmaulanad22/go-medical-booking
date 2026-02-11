-- Rollback: Remove queue_number from bookings table

DROP INDEX IF EXISTS idx_bookings_schedule_queue;
ALTER TABLE bookings DROP COLUMN IF EXISTS queue_number;
