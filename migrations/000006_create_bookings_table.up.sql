-- Migration: Create bookings table (Transaction)
-- Description: Stores patient booking transactions with status tracking

-- Create booking status enum type
CREATE TYPE booking_status AS ENUM ('pending', 'confirmed', 'cancelled');

CREATE TABLE IF NOT EXISTS bookings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    patient_id UUID NOT NULL REFERENCES patient_profiles(user_id) ON DELETE RESTRICT,
    schedule_id INTEGER NOT NULL REFERENCES doctor_schedules(id) ON DELETE RESTRICT,
    booking_code VARCHAR(50) NOT NULL UNIQUE,
    status booking_status NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Index for booking code lookups
CREATE UNIQUE INDEX IF NOT EXISTS idx_bookings_booking_code ON bookings(booking_code);

-- Index for patient booking history
CREATE INDEX IF NOT EXISTS idx_bookings_patient_id ON bookings(patient_id);

-- Index for schedule bookings
CREATE INDEX IF NOT EXISTS idx_bookings_schedule_id ON bookings(schedule_id);

-- Index for status-based queries
CREATE INDEX IF NOT EXISTS idx_bookings_status ON bookings(status);

-- Composite index for patient booking history with status filter
CREATE INDEX IF NOT EXISTS idx_bookings_patient_status ON bookings(patient_id, status);

COMMENT ON TABLE bookings IS 'Patient booking transactions - use optimistic locking for high-concurrency';
COMMENT ON COLUMN bookings.booking_code IS 'Unique human-readable booking reference code';
COMMENT ON COLUMN bookings.status IS 'pending = awaiting confirmation, confirmed = booking active, cancelled = booking cancelled';
