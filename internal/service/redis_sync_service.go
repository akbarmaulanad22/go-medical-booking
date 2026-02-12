package service

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"go-template-clean-architecture/internal/domain/entity"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// =============================================================================
// Errors
// =============================================================================

// ErrQuotaFull is returned when schedule slot is fully booked
var ErrQuotaFull = errors.New("schedule quota is full")

// decrQuotaIncrQueueScript is a package-level Lua script.
// Redis Go client automatically uses EVALSHA (send SHA hash only) after the first call,
// instead of EVAL (send full script text every time). This is significant for high-concurrency.
//
// Logic:
// 1. DECR quota key
// 2. If result < 0 → INCR back (rollback) and return -1 (quota full)
// 3. If result >= 0 → INCR queue key and return queue number
var decrQuotaIncrQueueScript = redis.NewScript(`
	local remaining = redis.call('DECR', KEYS[1])
	if remaining < 0 then
		redis.call('INCR', KEYS[1])
		return -1
	end
	local queue = redis.call('INCR', KEYS[2])
	return queue
`)

// =============================================================================
// Constants
// =============================================================================

const (
	// Redis key prefixes for booking system
	RedisQuotaKeyPrefix = "schedule:quota:"
	RedisQueueKeyPrefix = "booking:queue:"

	// Timeout for individual Redis operations
	redisSyncTimeout = 5 * time.Second

	// Batch size for startup sync - process 500 records at a time
	// CRITICAL: Pipeline is created and executed INSIDE the batch loop
	syncBatchSize = 500

	// Interval for cleaning up stale mutexes
	mutexCleanupInterval = 10 * time.Minute

	// How long a mutex must be unused before cleanup
	mutexStaleThreshold = 10 * time.Minute
)

// =============================================================================
// Types
// =============================================================================

// RedisSyncService handles syncing PostgreSQL data to Redis for the booking system.
//
// Key Features:
// - Memory Safe: Executes pipeline per batch (500 records) to prevent OOM
// - Concurrency Safe: Per-schedule mutex prevents race conditions
// - Atomic Operations: Uses Redis transactions for consistency
//
// Lock Ordering (to prevent deadlocks):
// 1. Acquire schedule mutex FIRST
// 2. Then perform DB/Redis operations
type RedisSyncService struct {
	db          *gorm.DB
	redisClient *redis.Client
	log         *logrus.Logger

	// Per-schedule mutex for concurrent safety
	scheduleMu sync.Map // map[int]*mutexWithTimestamp

	// Graceful shutdown
	stopChan chan struct{}
	wg       sync.WaitGroup
	stopped  atomic.Bool
}

// mutexWithTimestamp tracks mutex usage for cleanup
type mutexWithTimestamp struct {
	mu       sync.Mutex
	lastUsed atomic.Int64 // Unix timestamp
}

// QuotaResult holds quota sync data from database
type QuotaResult struct {
	ScheduleID     int
	TotalQuota     int
	RemainingQuota int
	MaxQueueNumber int
	ScheduleDate   time.Time
}

// =============================================================================
// Constructor
// =============================================================================

// NewRedisSyncService creates a new RedisSyncService.
// Starts background goroutine for mutex cleanup.
// Call Stop() during graceful shutdown.
func NewRedisSyncService(db *gorm.DB, redisClient *redis.Client, log *logrus.Logger) *RedisSyncService {
	svc := &RedisSyncService{
		db:          db,
		redisClient: redisClient,
		log:         log,
		stopChan:    make(chan struct{}),
	}

	// Start background cleanup goroutine
	svc.wg.Add(1)
	go svc.cleanupMutexMapLoop()

	return svc
}

// =============================================================================
// Lifecycle Methods
// =============================================================================

// Stop gracefully shuts down the service.
// Safe to call multiple times.
func (s *RedisSyncService) Stop() {
	if s.stopped.CompareAndSwap(false, true) {
		close(s.stopChan)
		s.wg.Wait()
		s.log.Info("RedisSyncService stopped")
	}
}

// =============================================================================
// Public Methods
// =============================================================================

// SyncOnStartup performs full sync of all active schedules from PostgreSQL to Redis.
//
// CRITICAL Fixes:
// - Calculates MAX(queue_number) from bookings table (not reset to 0)
// - Processes records in batches of 500
// - Creates and executes NEW pipeline INSIDE each batch loop
//
// Should be called BEFORE accepting traffic (during startup/disaster recovery).
func (s *RedisSyncService) SyncOnStartup(ctx context.Context) error {
	s.log.Info("Starting Redis re-sync from database...")
	startTime := time.Now()

	// Check Redis availability
	if err := s.redisClient.Ping(ctx).Err(); err != nil {
		s.log.Warnf("Redis is not available, skipping sync: %+v", err)
		return fmt.Errorf("redis ping failed: %w", err)
	}

	today := time.Now().UTC().Truncate(24 * time.Hour)
	offset := 0
	totalSynced := 0

	for {
		var results []QuotaResult

		// Batch query: get schedules with calculated remaining quota AND max queue number
		// CRITICAL FIX: Calculate MAX(queue_number) from bookings, not reset to 0
		err := s.db.Model(&entity.DoctorSchedule{}).
			Select(`
				doctor_schedules.id as schedule_id,
				doctor_schedules.total_quota,
				doctor_schedules.total_quota - COUNT(CASE WHEN bookings.status IS NOT NULL AND bookings.status != ? THEN 1 END) as remaining_quota,
				COALESCE(MAX(bookings.queue_number), 0) as max_queue_number,
				doctor_schedules.schedule_date
			`, string(entity.BookingStatusCancelled)).
			Joins("LEFT JOIN bookings ON bookings.schedule_id = doctor_schedules.id").
			Where("doctor_schedules.schedule_date >= ?", today).
			Group("doctor_schedules.id, doctor_schedules.total_quota, doctor_schedules.schedule_date").
			Order("doctor_schedules.id").
			Limit(syncBatchSize).
			Offset(offset).
			Scan(&results).Error

		if err != nil {
			s.log.Errorf("Failed to query schedules at offset %d: %+v", offset, err)
			return fmt.Errorf("query schedules at offset %d: %w", offset, err)
		}

		if len(results) == 0 {
			if offset == 0 {
				s.log.Info("No active schedules found for sync")
			}
			break
		}

		s.log.Infof("Processing batch: offset=%d, count=%d", offset, len(results))

		// CRITICAL: Create NEW pipeline for THIS batch only
		// This prevents memory accumulation across batches
		pipe := s.redisClient.TxPipeline()

		for _, result := range results {
			quotaKey := fmt.Sprintf("%s%d", RedisQuotaKeyPrefix, result.ScheduleID)
			queueKey := fmt.Sprintf("%s%d", RedisQueueKeyPrefix, result.ScheduleID)
			ttl := s.calculateTTL(result.ScheduleDate)

			// SET quota key (always overwrite with current DB value)
			pipe.Set(ctx, quotaKey, result.RemainingQuota, ttl)

			// SET queue key with MAX(queue_number) from DB
			// CRITICAL FIX: Use actual max queue number, not 0
			pipe.Set(ctx, queueKey, result.MaxQueueNumber, ttl)
		}

		// Execute pipeline for THIS batch
		if _, err := pipe.Exec(ctx); err != nil {
			s.log.Errorf("Failed to execute pipeline for batch at offset %d: %+v", offset, err)
			return fmt.Errorf("pipeline exec at offset %d: %w", offset, err)
		}

		totalSynced += len(results)
		s.log.Debugf("Synced batch: %d schedules", len(results))

		// Check if we've processed all records
		if len(results) < syncBatchSize {
			break
		}

		offset += syncBatchSize

		// Respect context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
	}

	elapsed := time.Since(startTime)
	s.log.Infof("Redis re-sync completed: %d schedules synced in %v", totalSynced, elapsed)

	return nil
}

// SyncScheduleQuota syncs a single schedule to Redis.
// Calculates remaining_quota from DB: TotalQuota - Count(non-cancelled bookings)
// Calculates max_queue_number from DB: MAX(queue_number) from bookings
//
// Called by: CreateSchedule (initial sync)
//
// Uses:
// - Redis Transaction for atomicity
// - Per-schedule mutex for concurrency safety
func (s *RedisSyncService) SyncScheduleQuota(ctx context.Context, scheduleID int, totalQuota int, scheduleDate time.Time) error {
	// Acquire per-schedule mutex
	mt := s.getScheduleMutex(scheduleID)
	mt.mu.Lock()
	defer mt.mu.Unlock()

	today := time.Now().UTC().Truncate(24 * time.Hour)

	// Skip past dates
	if scheduleDate.Before(today) {
		s.log.Debugf("Skipping sync for past schedule %d", scheduleID)
		return nil
	}

	// Query both booked count and max queue number in single query
	type syncData struct {
		BookedCount    int64
		MaxQueueNumber int
	}
	var data syncData

	err := s.db.WithContext(ctx).Model(&entity.Booking{}).
		Select("COUNT(*) as booked_count, COALESCE(MAX(queue_number), 0) as max_queue_number").
		Where("schedule_id = ? AND status != ?", scheduleID, entity.BookingStatusCancelled).
		Scan(&data).Error

	if err != nil {
		s.log.Warnf("Failed to query booking data for schedule %d: %+v", scheduleID, err)
		return fmt.Errorf("query booking data for schedule %d: %w", scheduleID, err)
	}

	remainingQuota := totalQuota - int(data.BookedCount)
	if remainingQuota < 0 {
		remainingQuota = 0
	}

	quotaKey := fmt.Sprintf("%s%d", RedisQuotaKeyPrefix, scheduleID)
	queueKey := fmt.Sprintf("%s%d", RedisQueueKeyPrefix, scheduleID)
	ttl := s.calculateTTL(scheduleDate)

	// Use Redis transaction for atomic operations
	pipe := s.redisClient.TxPipeline()

	// SET quota with TTL
	pipe.Set(ctx, quotaKey, remainingQuota, ttl)

	// SET queue with actual max from DB (not 0)
	pipe.Set(ctx, queueKey, data.MaxQueueNumber, ttl)

	if _, err := pipe.Exec(ctx); err != nil {
		s.log.Warnf("Failed to sync Redis for schedule %d: %+v", scheduleID, err)
		return fmt.Errorf("redis sync for schedule %d: %w", scheduleID, err)
	}

	s.log.Debugf("Synced schedule %d: quota=%d, queue=%d, TTL=%v", scheduleID, remainingQuota, data.MaxQueueNumber, ttl)
	return nil
}

// UpdateScheduleQuotaDelta updates Redis quota using INCRBY with delta.
//
// Called by: UpdateSchedule when TotalQuota changes
//
// Delta Strategy:
// - If Admin changes TotalQuota from 10 to 15, delta = +5
// - Uses INCRBY to atomically add delta to current Redis value
// - Includes bounds validation to prevent negative quotas
//
// Returns error if delta would result in negative quota
func (s *RedisSyncService) UpdateScheduleQuotaDelta(ctx context.Context, scheduleID int, delta int, scheduleDate time.Time) error {
	// Acquire per-schedule mutex
	mt := s.getScheduleMutex(scheduleID)
	mt.mu.Lock()
	defer mt.mu.Unlock()

	today := time.Now().UTC().Truncate(24 * time.Hour)

	// Skip past dates
	if scheduleDate.Before(today) {
		s.log.Debugf("Skipping delta update for past schedule %d", scheduleID)
		return nil
	}

	quotaKey := fmt.Sprintf("%s%d", RedisQuotaKeyPrefix, scheduleID)
	ttl := s.calculateTTL(scheduleDate)

	// BOUNDS VALIDATION: Check current quota before applying negative delta
	if delta < 0 {
		currentQuota, err := s.redisClient.Get(ctx, quotaKey).Int()
		if err != nil && err != redis.Nil {
			s.log.Warnf("Failed to get current quota for schedule %d: %+v", scheduleID, err)
			return fmt.Errorf("get current quota for schedule %d: %w", scheduleID, err)
		}

		// Prevent negative quota
		if currentQuota+delta < 0 {
			s.log.Warnf("Delta %d would result in negative quota (current: %d) for schedule %d", delta, currentQuota, scheduleID)
			// Adjust delta to set quota to 0 instead of negative
			delta = -currentQuota
		}
	}

	// Use INCRBY for atomic delta update
	pipe := s.redisClient.TxPipeline()
	pipe.IncrBy(ctx, quotaKey, int64(delta))
	pipe.Expire(ctx, quotaKey, ttl)

	if _, err := pipe.Exec(ctx); err != nil {
		s.log.Warnf("Failed to update quota delta for schedule %d: %+v", scheduleID, err)
		return fmt.Errorf("update quota delta for schedule %d: %w", scheduleID, err)
	}

	s.log.Debugf("Updated schedule %d quota by delta=%d", scheduleID, delta)
	return nil
}

// DeleteScheduleKeys removes quota and queue keys from Redis.
// Also immediately cleans up the mutex from memory.
//
// Called by: DeleteSchedule after successful DB deletion
func (s *RedisSyncService) DeleteScheduleKeys(ctx context.Context, scheduleID int) error {
	// Acquire per-schedule mutex
	mt := s.getScheduleMutex(scheduleID)
	mt.mu.Lock()
	defer func() {
		mt.mu.Unlock()
		// IMMEDIATE CLEANUP: Remove mutex from map after delete
		s.scheduleMu.Delete(scheduleID)
	}()

	quotaKey := fmt.Sprintf("%s%d", RedisQuotaKeyPrefix, scheduleID)
	queueKey := fmt.Sprintf("%s%d", RedisQueueKeyPrefix, scheduleID)

	if err := s.redisClient.Del(ctx, quotaKey, queueKey).Err(); err != nil {
		s.log.Warnf("Failed to delete Redis keys for schedule %d: %+v", scheduleID, err)
		return fmt.Errorf("delete redis keys for schedule %d: %w", scheduleID, err)
	}

	s.log.Debugf("Deleted Redis keys for schedule %d", scheduleID)
	return nil
}

// DecrQuotaAndIncrQueue atomically reserves a booking slot and gets queue number.
//
// HIGH CONCURRENCY STRATEGY — Lua Script:
// Executes DECR quota + INCR queue as a SINGLE atomic operation inside Redis.
// No failure window between the two steps — either both succeed or neither applies.
//
// NO MUTEX NEEDED: Lua scripts execute atomically in Redis (single-threaded).
// An in-app mutex would serialize all requests per schedule, becoming a bottleneck.
//
// Called by: CreateBooking usecase
//
// Returns: queue number (1-based), or error
func (s *RedisSyncService) DecrQuotaAndIncrQueue(ctx context.Context, scheduleID int) (int, error) {
	quotaKey := fmt.Sprintf("%s%d", RedisQuotaKeyPrefix, scheduleID)
	queueKey := fmt.Sprintf("%s%d", RedisQueueKeyPrefix, scheduleID)

	// Uses package-level decrQuotaIncrQueueScript for EVALSHA optimization
	result, err := decrQuotaIncrQueueScript.Run(ctx, s.redisClient, []string{quotaKey, queueKey}).Int()
	if err != nil {
		s.log.Warnf("Failed Lua script DecrQuotaAndIncrQueue for schedule %d: %+v", scheduleID, err)
		return 0, fmt.Errorf("lua decrquota_incrqueue for schedule %d: %w", scheduleID, err)
	}

	if result == -1 {
		return 0, ErrQuotaFull
	}

	s.log.Debugf("Reserved slot for schedule %d: queue_number=%d", scheduleID, result)
	return result, nil
}

// RestoreQuota restores a booking slot when a booking is cancelled.
//
// IMPORTANT: Only increments quota, does NOT decrement queue number.
// Queue numbers are monotonically increasing and never reused.
//
// Called by: CancelBooking usecase
func (s *RedisSyncService) RestoreQuota(ctx context.Context, scheduleID int) error {
	// Acquire per-schedule mutex
	mt := s.getScheduleMutex(scheduleID)
	mt.mu.Lock()
	defer mt.mu.Unlock()

	quotaKey := fmt.Sprintf("%s%d", RedisQuotaKeyPrefix, scheduleID)

	if err := s.redisClient.Incr(ctx, quotaKey).Err(); err != nil {
		s.log.Warnf("Failed to restore quota for schedule %d: %+v", scheduleID, err)
		return fmt.Errorf("restore quota for schedule %d: %w", scheduleID, err)
	}

	s.log.Debugf("Restored quota for schedule %d (cancel)", scheduleID)
	return nil
}

// =============================================================================
// Private Helper Methods
// =============================================================================

// getScheduleMutex returns mutex for a specific schedule ID
func (s *RedisSyncService) getScheduleMutex(scheduleID int) *mutexWithTimestamp {
	mt, _ := s.scheduleMu.LoadOrStore(scheduleID, &mutexWithTimestamp{})
	result := mt.(*mutexWithTimestamp)
	result.lastUsed.Store(time.Now().Unix())
	return result
}

// cleanupMutexMapLoop runs in background to clean stale mutexes
func (s *RedisSyncService) cleanupMutexMapLoop() {
	defer s.wg.Done()

	ticker := time.NewTicker(mutexCleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			s.log.Debug("Mutex cleanup goroutine stopping")
			return
		case <-ticker.C:
			s.cleanupStaleMutexes()
		}
	}
}

// cleanupStaleMutexes removes unused mutexes using TryLock for safety
// FIXED: Moved lastUsed check inside lock to prevent race condition
func (s *RedisSyncService) cleanupStaleMutexes() {
	cutoffTime := time.Now().Add(-mutexStaleThreshold).Unix()
	var cleaned int

	s.scheduleMu.Range(func(key, value any) bool {
		scheduleID, ok := key.(int)
		if !ok {
			return true
		}

		mt, ok := value.(*mutexWithTimestamp)
		if !ok {
			return true
		}

		// TryLock first - if we can't get lock, someone is using it
		if mt.mu.TryLock() {
			// FIXED: Check lastUsed INSIDE lock to prevent race condition
			// This ensures no one updated lastUsed between our check and lock
			if mt.lastUsed.Load() < cutoffTime {
				s.scheduleMu.Delete(scheduleID)
				cleaned++
			}
			mt.mu.Unlock()
		}
		return true
	})

	if cleaned > 0 {
		s.log.Debugf("Cleaned up %d stale mutexes", cleaned)
	}
}

// calculateTTL returns TTL: 24 hours after schedule date
func (s *RedisSyncService) calculateTTL(scheduleDate time.Time) time.Duration {
	expireAt := scheduleDate.AddDate(0, 0, 1)
	ttl := time.Until(expireAt)

	if ttl <= 0 {
		// Past date - short TTL for cleanup
		return 1 * time.Minute
	}

	return ttl
}
