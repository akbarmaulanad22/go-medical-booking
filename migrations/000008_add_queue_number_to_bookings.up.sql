-- Migration: Add queue_number to bookings table
-- Description: Stores the patient's queue position for a schedule slot

ALTER TABLE bookings ADD COLUMN queue_number INTEGER NOT NULL DEFAULT 0;

-- Composite index for queue number lookup per schedule
CREATE INDEX IF NOT EXISTS idx_bookings_schedule_queue ON bookings(schedule_id, queue_number);
