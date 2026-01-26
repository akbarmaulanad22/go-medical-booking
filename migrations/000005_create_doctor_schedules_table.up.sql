-- Migration: Create doctor_schedules table (Quota Management)
-- Description: Stores doctor availability schedules with quota tracking for bookings

CREATE TABLE IF NOT EXISTS doctor_schedules (
    id SERIAL PRIMARY KEY,
    doctor_id UUID NOT NULL REFERENCES doctor_profiles(user_id) ON DELETE CASCADE,
    schedule_date DATE NOT NULL,
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    total_quota INTEGER NOT NULL CHECK (total_quota > 0),
    remaining_quota INTEGER NOT NULL CHECK (remaining_quota >= 0),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_remaining_quota CHECK (remaining_quota <= total_quota),
    CONSTRAINT chk_time_order CHECK (end_time > start_time)
);

-- Index for doctor schedule lookups
CREATE INDEX IF NOT EXISTS idx_doctor_schedules_doctor_id ON doctor_schedules(doctor_id);

-- Index for date-based queries
CREATE INDEX IF NOT EXISTS idx_doctor_schedules_schedule_date ON doctor_schedules(schedule_date);

-- Composite index for availability searches
CREATE INDEX IF NOT EXISTS idx_doctor_schedules_available ON doctor_schedules(doctor_id, schedule_date, remaining_quota) 
    WHERE remaining_quota > 0;

COMMENT ON TABLE doctor_schedules IS 'Doctor schedule slots with quota management for high-concurrency bookings';
COMMENT ON COLUMN doctor_schedules.remaining_quota IS 'Decremented atomically during booking, use FOR UPDATE in transactions';
