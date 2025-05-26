package timestate

import (
	"context"
	"log"
	"sync"
	"sync/atomic"
	"time"
	"fmt"
)

type callbackManager struct {
	mu        sync.RWMutex
	callbacks []func(time.Time)
}

// 修改为支持 context 的回调签名
type TimedCallbackFunc func(ctx context.Context, t time.Time)

type timedCallback struct {
	Hour, Minute int
	Callback     TimedCallbackFunc
}

type timeKey struct {
	Hour, Minute int
}
var (
	currentWeekKey atomic.Value
	nextWeekKey    atomic.Value
)

var callbackMap atomic.Value

var (
	zeroMsTimestamp    int64        // 当天零点的毫秒时间戳
	zeroSecTimestamp   int64        // 当天零点的秒时间戳
	zeroDateValue      atomic.Value // 当天日期字符串

	nextDayTimestamp   int64
	nextWeekTimestamp  int64
	nextMonthTimestamp int64

	dayCallbacks   = &callbackManager{}
	weekCallbacks  = &callbackManager{}
	monthCallbacks = &callbackManager{}

	once sync.Once
)

var (
    dailyTimedMu        sync.RWMutex
    dailyTimedCallbacks []timedCallback
    dailyTimedRegisterChan = make(chan struct{}, 1)
)


func InitTimezoneTimer(tz *time.Location) {
	once.Do(func() {
// 		tz := time.Local // Use local timezone by default
		now := time.Now().In(tz)
		midnight := getTodayMidnight(now)

		atomic.StoreInt64(&zeroMsTimestamp, midnight.UnixMilli())
		atomic.StoreInt64(&zeroSecTimestamp, midnight.Unix())
		zeroDateValue.Store(midnight.Format("2006-01-02"))

		updateTimeStates(midnight)
		go startMidnightTimer(tz)
		go startDailyTimers(tz)
	})
}

// startMidnightTimer 每天零点触发一次
func startMidnightTimer(tz *time.Location) {
	for {
		now := time.Now().In(tz)

		next := getTodayMidnight(now).Add(24 * time.Hour)
		duration := next.Sub(now)

		log.Printf("[timestate] 下次 0 点触发时间: %v (%.1f小时后)", next.Format("2006-01-02 15:04:05"), duration.Hours())
		timer := time.NewTimer(duration)
		<-timer.C
		now = time.Now().In(tz)
		nowTs := now.Unix()

		prevDay := atomic.LoadInt64(&zeroSecTimestamp)
		nextDay := atomic.LoadInt64(&nextDayTimestamp)
		nextWeek := atomic.LoadInt64(&nextWeekTimestamp)
		nextMonth := atomic.LoadInt64(&nextMonthTimestamp)

		updateTimeStates(now)

		go checkAndTriggerCallbacks(now, nowTs, prevDay, nextDay, nextWeek, nextMonth)
	}
}

func checkAndTriggerCallbacks(now time.Time, nowTs, prevDay, nextDay, nextWeek, nextMonth int64) {
	log.Printf("[timestate] 当前时间戳: %d", nowTs)
	log.Printf("[timestate] 零点时间戳: %d", prevDay)
	log.Printf("[timestate] 下一天时间戳: %d", nextDay)
	log.Printf("[timestate] 下一周时间戳: %d", nextWeek)
	log.Printf("[timestate] 下个月时间戳: %d", nextMonth)

	if nowTs >= nextDay {
		log.Printf("[timestate] 触发新的一天回调 @ %v", now.Format("2006-01-02"))
		dayCallbacks.triggerCallbacks(now)
	}

	if nowTs >= nextWeek {
		log.Printf("[timestate] 触发新的一周回调 @ %v", now.Format("2006-01-02"))
		weekCallbacks.triggerCallbacks(now)
	}

	if nowTs >= nextMonth {
		log.Printf("[timestate] 触发新的一月回调 @ %v", now.Format("2006-01"))
		monthCallbacks.triggerCallbacks(now)
	}
}

func startDailyTimers(tz *time.Location) {
	callbackMap.Store(make(map[timeKey][]TimedCallbackFunc))

	defer log.Println("[timestate] startDailyTimers finish")

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

	rebuildCallbackMap()

	// 异步监听注册变更
	go func() {
		for range dailyTimedRegisterChan {
			rebuildCallbackMap()
		}
	}()

	// 定时触发器
	go func() {
		for {
			now := time.Now().In(tz)
			nextTick := now.Truncate(time.Minute).Add(time.Minute)
			time.Sleep(time.Until(nextTick)) // 等到下一个整分钟

			now = nextTick
			key := timeKey{now.Hour(), now.Minute()}

			cbMap, ok := callbackMap.Load().(map[timeKey][]TimedCallbackFunc)
			if !ok {
				log.Println("[timestate] 错误: callback map 类型断言失败")
				continue
			}
			callbacks := cbMap[key]
			if len(callbacks) == 0 {
				continue
			}

			log.Printf("[timestate] 精准触发 %d 个每日定时任务 @ %02d:%02d", len(callbacks), now.Hour(), now.Minute())

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
				// 正常完成
			case <-ctx.Done():
				log.Printf("[timestate] 警告: 每日定时任务执行超时 @ %02d:%02d", now.Hour(), now.Minute())
			}
			cancel()
		}
	}()
}

// updateTimeStates 更新零点与下周期时间戳
func updateTimeStates(t time.Time) {
	today := getTodayMidnight(t)
	currentKey := getCurrentPeriodKey(t)
	nextKey := getNextPeriodKey(t)

	currentWeekKey.Store(currentKey)
	nextWeekKey.Store(nextKey)
	atomic.StoreInt64(&zeroMsTimestamp, today.UnixMilli())
	atomic.StoreInt64(&zeroSecTimestamp, today.Unix())
	zeroDateValue.Store(today.Format("2006-01-02"))

	atomic.StoreInt64(&nextDayTimestamp, today.Add(24*time.Hour).Unix())

	offset := (8 - int(today.Weekday())) % 7
	if offset == 0 {
		offset = 7
	}
	nextWeek := today.AddDate(0, 0, offset)
	atomic.StoreInt64(&nextWeekTimestamp, nextWeek.Unix())

	nextMonth := time.Date(today.Year(), today.Month()+1, 1, 0, 0, 0, 0, today.Location())
	atomic.StoreInt64(&nextMonthTimestamp, nextMonth.Unix())
}

func getTodayMidnight(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// BuildPeriodKey 生成指定时间的周期 Key，格式为 "YYYY-MM-Wn"
func BuildPeriodKey(t time.Time) string {
	year, month := t.Year(), int(t.Month())
	week := GetWeekOfMonth(t)
	return fmt.Sprintf("%04d-%02d-W%d", year, month, week)
}

// GetCurrentPeriodKey 返回当前时间的周期 Key
func getCurrentPeriodKey(t time.Time) string {
	return BuildPeriodKey(t)
}

// GetNextPeriodKey 返回当前时间 +7 天后的周期 Key
func getNextPeriodKey(t time.Time) string {
	return BuildPeriodKey(t.AddDate(0, 0, 7))
}

// GetWeekOfMonth 根据给定时间，返回该时间在本月的第几周（周一为起始）
func GetWeekOfMonth(t time.Time) int {
    // 调整星期，将周一变为0，周日变为6
    adjustedWeekday := func(day time.Weekday) int {
        return (int(day) + 6) % 7
    }

    firstDay := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
    offsetDays := t.Day() - 1 + adjustedWeekday(firstDay.Weekday())

    // 计算第几周，整除7后加1
    return (offsetDays / 7) + 1
}


func (m *callbackManager) triggerCallbacks(t time.Time) {
	m.mu.RLock()
	cbs := append([]func(time.Time){}, m.callbacks...)
	m.mu.RUnlock()

	for _, cb := range cbs {
		go safeCall(cb, t)
	}
}

func safeCall(cb func(time.Time), t time.Time) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[timestate] 回调 panic: %v", r)
		}
	}()
	cb(t)
}

// ===========================
// ✅ 注册接口（改造每日定时回调）
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

// 注册接口改成带 context 参数
func RegisterDailyTimeCallback(hour, minute int, fn TimedCallbackFunc) {
    if hour < 0 || hour > 23 || minute < 0 || minute > 59 {
        log.Printf("[timestate] 无效时间参数 hour=%d, minute=%d", hour, minute)
        return
    }
    dailyTimedMu.Lock()
    dailyTimedCallbacks = append(dailyTimedCallbacks, timedCallback{
        Hour:     hour,
        Minute:   minute,
        Callback: fn,
    })
    dailyTimedMu.Unlock()

    select {
    case dailyTimedRegisterChan <- struct{}{}:
    default:
    }
}

// 触发指定时间回调，带超时控制
func TriggerDailyCallbackWithTimeout(hour, minute int, timeout time.Duration) {
    dailyTimedMu.RLock()
    defer dailyTimedMu.RUnlock()

    var matched []TimedCallbackFunc
    for _, tcb := range dailyTimedCallbacks {
        if tcb.Hour == hour && tcb.Minute == minute {
            matched = append(matched, tcb.Callback)
        }
    }

    if len(matched) == 0 {
        log.Printf("[timestate] 没有找到匹配的定时回调 @ %02d:%02d", hour, minute)
        return
    }

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
        log.Printf("[timestate] 所有定时回调执行完成 @ %02d:%02d", hour, minute)
    case <-ctx.Done():
        log.Printf("[timestate] 定时回调执行超时 @ %02d:%02d", hour, minute)
    }
}


func safeCallContext(cb TimedCallbackFunc, ctx context.Context, t time.Time) {
    defer func() {
        if r := recover(); r != nil {
            log.Printf("[timestate] 带 context 的回调 panic: %v", r)
        }
    }()
    cb(ctx, t)
}

// ===========================
// ✅ 读取接口
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