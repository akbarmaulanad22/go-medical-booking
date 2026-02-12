package entity

// ScheduleFilter is a domain-level filter for querying schedules.
// Used by repository layer to avoid coupling with delivery DTOs.
type ScheduleFilter struct {
	StartAt        string // Format: YYYY-MM-DD
	EndAt          string // Format: YYYY-MM-DD
	DoctorName     string // Filter by doctor name (ILIKE)
	Specialization string // Filter by specialization (ILIKE)
}
