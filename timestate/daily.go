package timestate

import (
	"context"
	"log"
	"sync"
	"sync/atomic"
	"time"
	"fmt"
)

// callbackManager manages a collection of time-based callbacks with thread-safe access
type callbackManager struct {
	mu        sync.RWMutex
	callbacks []func(time.Time)
}

// TimedCallbackFunc defines the signature for time-based callbacks with context support
type TimedCallbackFunc func(ctx context.Context, t time.Time)

// timedCallback represents a scheduled daily callback at specific hour and minute
type timedCallback struct {
	Hour, Minute int
	Callback     TimedCallbackFunc
}

// timeKey is used as a map key for storing timed callbacks
type timeKey struct {
	Hour, Minute int
}

var (
	// Atomic storage for current and next week period keys
	currentWeekKey atomic.Value
	nextWeekKey    atomic.Value
)

// Atomic storage for the callback map
var callbackMap atomic.Value

var (
	// Timestamps for the current day's midnight (millisecond and second precision)
	zeroMsTimestamp    int64
	zeroSecTimestamp   int64
	zeroDateValue      atomic.Value // Current date as string in "YYYY-MM-DD" format

	// Timestamps for next day/week/month boundaries
	nextDayTimestamp   int64
	nextWeekTimestamp  int64
	nextMonthTimestamp int64

	// Callback managers for different time periods
	dayCallbacks   = &callbackManager{}
	weekCallbacks  = &callbackManager{}
	monthCallbacks = &callbackManager{}

	// Ensures initialization happens only once
	once sync.Once
)

var (
    // Synchronization for daily timed callbacks
    dailyTimedMu        sync.RWMutex
    dailyTimedCallbacks []timedCallback
    dailyTimedRegisterChan = make(chan struct{}, 1) // Notification channel for callback registration
)

// InitTimezoneTimer initializes the time state tracking system with the specified timezone
func InitTimezoneTimer(tz *time.Location) {
	once.Do(func() {
		now := time.Now().In(tz)
		midnight := getTodayMidnight(now)

		// Initialize atomic time states
		atomic.StoreInt64(&zeroMsTimestamp, midnight.UnixMilli())
		atomic.StoreInt64(&zeroSecTimestamp, midnight.Unix())
		zeroDateValue.Store(midnight.Format("2006-01-02"))

		updateTimeStates(midnight)
		go startMidnightTimer(tz)
		go startDailyTimers(tz)
	})
}

// startMidnightTimer runs a loop that triggers at each midnight to update time states
func startMidnightTimer(tz *time.Location) {
	for {
		now := time.Now().In(tz)

		// Calculate next midnight and sleep until then
		next := getTodayMidnight(now).Add(24 * time.Hour)
		duration := next.Sub(now)

		log.Printf("[timestate] Next midnight trigger: %v (in %.1f hours)", next.Format("2006-01-02 15:04:05"), duration.Hours())
		timer := time.NewTimer(duration)
		<-timer.C
		now = time.Now().In(tz)
		nowTs := now.Unix()

		// Load current time states
		prevDay := atomic.LoadInt64(&zeroSecTimestamp)
		nextDay := atomic.LoadInt64(&nextDayTimestamp)
		nextWeek := atomic.LoadInt64(&nextWeekTimestamp)
		nextMonth := atomic.LoadInt64(&nextMonthTimestamp)

		// Update all time states
		updateTimeStates(now)

		// Check and trigger appropriate callbacks
		go checkAndTriggerCallbacks(now, nowTs, prevDay, nextDay, nextWeek, nextMonth)
	}
}

// checkAndTriggerCallbacks determines which periodic callbacks need to be triggered
func checkAndTriggerCallbacks(now time.Time, nowTs, prevDay, nextDay, nextWeek, nextMonth int64) {
	log.Printf("[timestate] Current timestamp: %d", nowTs)
	log.Printf("[timestate] Midnight timestamp: %d", prevDay)
	log.Printf("[timestate] Next day timestamp: %d", nextDay)
	log.Printf("[timestate] Next week timestamp: %d", nextWeek)
	log.Printf("[timestate] Next month timestamp: %d", nextMonth)

	if nowTs >= nextDay {
		log.Printf("[timestate] Triggering new day callbacks @ %v", now.Format("2006-01-02"))
		dayCallbacks.triggerCallbacks(now)
	}

	if nowTs >= nextWeek {
		log.Printf("[timestate] Triggering new week callbacks @ %v", now.Format("2006-01-02"))
		weekCallbacks.triggerCallbacks(now)
	}

	if nowTs >= nextMonth {
		log.Printf("[timestate] Triggering new month callbacks @ %v", now.Format("2006-01"))
		monthCallbacks.triggerCallbacks(now)
	}
}

// startDailyTimers manages precise daily callbacks at specific times
func startDailyTimers(tz *time.Location) {
	// Initialize empty callback map
	callbackMap.Store(make(map[timeKey][]TimedCallbackFunc))

	defer log.Println("[timestate] Daily timer service started")

	// rebuildCallbackMap reconstructs the callback lookup map from registered callbacks
	rebuildCallbackMap := func() {
		dailyTimedMu.RLock()
		defer dailyTimedMu.RUnlock()

		newMap := make(map[timeKey][]TimedCallbackFunc, len(dailyTimedCallbacks))
		for _, tcb := range dailyTimedCallbacks {
			key := timeKey{tcb.Hour, tcb.Minute}
			newMap[key] = append(newMap[key], tcb.Callback)
		}
		callbackMap.Store(newMap)
	}

	// Initial map build
	rebuildCallbackMap()

	// Watch for registration changes
	go func() {
		for range dailyTimedRegisterChan {
			rebuildCallbackMap()
		}
	}()

	// Main timing loop
	go func() {
		for {
			now := time.Now().In(tz)
			nextTick := now.Truncate(time.Minute).Add(time.Minute)
			time.Sleep(time.Until(nextTick)) // Wait until next minute

			now = nextTick
			key := timeKey{now.Hour(), now.Minute()}

			// Load and execute matching callbacks
			cbMap, ok := callbackMap.Load().(map[timeKey][]TimedCallbackFunc)
			if !ok {
				log.Println("[timestate] Error: callback map type assertion failed")
				continue
			}
			callbacks := cbMap[key]
			if len(callbacks) == 0 {
				continue
			}

			log.Printf("[timestate] Triggering %d daily callbacks @ %02d:%02d", len(callbacks), now.Hour(), now.Minute())

			// Execute callbacks with timeout control
			var wg sync.WaitGroup
			wg.Add(len(callbacks))
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

			for _, cb := range callbacks {
				go func(fn TimedCallbackFunc) {
					defer wg.Done()
					fn(ctx, now)
				}(cb)
			}

			done := make(chan struct{})
			go func() {
				wg.Wait()
				close(done)
			}()

			select {
			case <-done:
				// All callbacks completed
			case <-ctx.Done():
				log.Printf("[timestate] Warning: Daily callbacks timed out @ %02d:%02d", now.Hour(), now.Minute())
			}
			cancel()
		}
	}()
}

// updateTimeStates recalculates all time-based state variables
func updateTimeStates(t time.Time) {
	today := getTodayMidnight(t)
	currentKey := getCurrentPeriodKey(t)
	nextKey := getNextPeriodKey(t)

	// Update atomic values
	currentWeekKey.Store(currentKey)
	nextWeekKey.Store(nextKey)
	atomic.StoreInt64(&zeroMsTimestamp, today.UnixMilli())
	atomic.StoreInt64(&zeroSecTimestamp, today.Unix())
	zeroDateValue.Store(today.Format("2006-01-02"))

	// Calculate next boundaries
	atomic.StoreInt64(&nextDayTimestamp, today.Add(24*time.Hour).Unix())

	// Calculate next week (following Monday)
	offset := (8 - int(today.Weekday())) % 7
	if offset == 0 {
		offset = 7
	}
	nextWeek := today.AddDate(0, 0, offset)
	atomic.StoreInt64(&nextWeekTimestamp, nextWeek.Unix())

	// Calculate next month (first day)
	nextMonth := time.Date(today.Year(), today.Month()+1, 1, 0, 0, 0, 0, today.Location())
	atomic.StoreInt64(&nextMonthTimestamp, nextMonth.Unix())
}

// getTodayMidnight returns the midnight time for the given date
func getTodayMidnight(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// BuildPeriodKey generates a period key string in "YYYY-MM-Wn" format
func BuildPeriodKey(t time.Time) string {
	year, month := t.Year(), int(t.Month())
	week := GetWeekOfMonth(t)
	return fmt.Sprintf("%04d-%02d-W%d", year, month, week)
}

// getCurrentPeriodKey returns the period key for the current time
func getCurrentPeriodKey(t time.Time) string {
	return BuildPeriodKey(t)
}

// getNextPeriodKey returns the period key for the time 7 days from now
func getNextPeriodKey(t time.Time) string {
	return BuildPeriodKey(t.AddDate(0, 0, 7))
}

// GetWeekOfMonth calculates which week of the month the date falls in (Monday-based)
func GetWeekOfMonth(t time.Time) int {
    // Adjust weekday so Monday=0, Sunday=6
    adjustedWeekday := func(day time.Weekday) int {
        return (int(day) + 6) % 7
    }

    firstDay := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
    offsetDays := t.Day() - 1 + adjustedWeekday(firstDay.Weekday())

    // Calculate week number (1-based)
    return (offsetDays / 7) + 1
}

// triggerCallbacks safely executes all registered callbacks
func (m *callbackManager) triggerCallbacks(t time.Time) {
	m.mu.RLock()
	cbs := append([]func(time.Time){}, m.callbacks...)
	m.mu.RUnlock()

	for _, cb := range cbs {
		go safeCall(cb, t)
	}
}

// safeCall executes a callback with panic recovery
func safeCall(cb func(time.Time), t time.Time) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[timestate] Callback panic: %v", r)
		}
	}()
	cb(t)
}

// ===========================
// ✅ Callback Registration
// ===========================

func RegisterDayCallback(fn func(time.Time)) {
	dayCallbacks.mu.Lock()
	defer dayCallbacks.mu.Unlock()
	dayCallbacks.callbacks = append(dayCallbacks.callbacks, fn)
}

func RegisterWeekCallback(fn func(time.Time)) {
	weekCallbacks.mu.Lock()
	defer weekCallbacks.mu.Unlock()
	weekCallbacks.callbacks = append(weekCallbacks.callbacks, fn)
}

func RegisterMonthCallback(fn func(time.Time)) {
	monthCallbacks.mu.Lock()
	defer monthCallbacks.mu.Unlock()
	monthCallbacks.callbacks = append(monthCallbacks.callbacks, fn)
}

// RegisterDailyTimeCallback adds a new timed daily callback
func RegisterDailyTimeCallback(hour, minute int, fn TimedCallbackFunc) {
    if hour < 0 || hour > 23 || minute < 0 || minute > 59 {
        log.Printf("[timestate] Invalid time parameters hour=%d, minute=%d", hour, minute)
        return
    }
    dailyTimedMu.Lock()
    dailyTimedCallbacks = append(dailyTimedCallbacks, timedCallback{
        Hour:     hour,
        Minute:   minute,
        Callback: fn,
    })
    dailyTimedMu.Unlock()

    // Notify about registration (non-blocking)
    select {
    case dailyTimedRegisterChan <- struct{}{}:
    default:
    }
}

// TriggerDailyCallbackWithTimeout manually triggers callbacks for a specific time with timeout control
func TriggerDailyCallbackWithTimeout(hour, minute int, timeout time.Duration) {
    dailyTimedMu.RLock()
    defer dailyTimedMu.RUnlock()

    // Find matching callbacks
    var matched []TimedCallbackFunc
    for _, tcb := range dailyTimedCallbacks {
        if tcb.Hour == hour && tcb.Minute == minute {
            matched = append(matched, tcb.Callback)
        }
    }

    if len(matched) == 0 {
        log.Printf("[timestate] No matching callbacks found @ %02d:%02d", hour, minute)
        return
    }

    // Execute with timeout
    ctx, cancel := context.WithTimeout(context.Background(), timeout)
    defer cancel()

    var wg sync.WaitGroup
    wg.Add(len(matched))

    for _, cb := range matched {
        go func(fn TimedCallbackFunc) {
            defer wg.Done()
            safeCallContext(fn, ctx, time.Now())
        }(cb)
    }

    done := make(chan struct{})
    go func() {
        wg.Wait()
        close(done)
    }()

    select {
    case <-done:
        log.Printf("[timestate] All callbacks completed @ %02d:%02d", hour, minute)
    case <-ctx.Done():
        log.Printf("[timestate] Callback execution timed out @ %02d:%02d", hour, minute)
    }
}

// safeCallContext executes a context-aware callback with panic recovery
func safeCallContext(cb TimedCallbackFunc, ctx context.Context, t time.Time) {
    defer func() {
        if r := recover(); r != nil {
            log.Printf("[timestate] Context callback panic: %v", r)
        }
    }()
    cb(ctx, t)
}

// ===========================
// ✅ State Accessors
// ===========================
func GetZeroMs() int64   { return atomic.LoadInt64(&zeroMsTimestamp) }
func GetZeroSec() int64  { return atomic.LoadInt64(&zeroSecTimestamp) }
func GetZeroDate() string {
	if s, ok := zeroDateValue.Load().(string); ok {
		return s
	}
	return ""
}
func GetZeroMonth() string {
	if s, ok := zeroDateValue.Load().(string); ok && len(s) >= 7 {
		return s[:7]
	}
	return ""
}
func GetCurrentPeriodKey() string {
	if val, ok := currentWeekKey.Load().(string); ok {
		return val
	}
	return ""
}

func GetNextPeriodKey() string {
	if val, ok := nextWeekKey.Load().(string); ok {
		return val
	}
	return ""
}
func GetNextDayTimestamp() int64   { return atomic.LoadInt64(&nextDayTimestamp) }
func GetNextWeekTimestamp() int64  { return atomic.LoadInt64(&nextWeekTimestamp) }
func GetNextMonthTimestamp() int64 { return atomic.LoadInt64(&nextMonthTimestamp) }