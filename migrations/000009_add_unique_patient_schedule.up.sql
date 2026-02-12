-- Migration: Add partial unique constraint to prevent duplicate active bookings
-- Description: Prevents race condition where two requests book the same schedule for same patient

-- Partial unique index: only enforced for non-cancelled bookings
-- This allows patients to re-book after cancelling
CREATE UNIQUE INDEX IF NOT EXISTS idx_bookings_patient_schedule_active
    ON bookings(patient_id, schedule_id)
    WHERE status != 'cancelled';
