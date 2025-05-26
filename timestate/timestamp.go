package timestate

import (
	"log"
	"sync/atomic"
	"time"
)

var (
	currentSecond      atomic.Int64 // 精确秒级时间戳(Unix秒)
	currentMillisecond atomic.Int64 // 对齐100ms边界的毫秒级时间戳
	lastUpdateTime     time.Time    // 记录上次更新时间(用于补偿)
)

// InitTimer 初始化精确100ms间隔计时器
// tz: 时区设置，nil表示使用本地时区
func InitTimer(tz *time.Location) {
	now := getCurrentTime(tz)
	alignedNow := alignTo100ms(now)
	
	currentMillisecond.Store(alignedNow.UnixMilli())
	currentSecond.Store(alignedNow.Unix())
	lastUpdateTime = alignedNow

	// 启动精确100ms间隔的更新器
	go runPrecise100msUpdater(tz)

	log.Printf("[timestate] 精确100ms间隔计时器已初始化 (当前对齐时间: %v)", alignedNow.Format("2006-01-02 15:04:05.000"))
}

// alignTo100ms 将时间对齐到最近的100ms边界
func alignTo100ms(t time.Time) time.Time {
	ms := t.UnixMilli()
	alignedMs := ms - (ms % 100)
	return time.UnixMilli(alignedMs)
}

// runPrecise100msUpdater 精确100ms间隔更新器
func runPrecise100msUpdater(tz *time.Location) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case now := <-ticker.C:
			// 计算补偿时间，确保精确100ms间隔
			expectedTime := lastUpdateTime.Add(100 * time.Millisecond)
			adjustedTime := now
			
			// 如果系统负载导致延迟，使用预期时间而非实际时间
			if now.Sub(expectedTime) > 5*time.Millisecond {
				adjustedTime = expectedTime
			}

			alignedTime := alignTo100ms(adjustedTime)
			ms := alignedTime.UnixMilli()
			
			currentMillisecond.Store(ms)
			if alignedTime.Unix() != lastUpdateTime.Unix() {
				currentSecond.Store(alignedTime.Unix())
			}
			
			lastUpdateTime = alignedTime
		}
	}
}

func getCurrentTime(tz *time.Location) time.Time {
	if tz == nil {
		return time.Now()
	}
	return time.Now().In(tz)
}

// GetSecond 获取当前精确秒级时间戳
func GetSecond() int64 {
	return currentSecond.Load()
}

// GetMillisecond 获取对齐100ms边界的毫秒级时间戳
func GetMillisecond() int64 {
	return currentMillisecond.Load()
}

// Timestamp 时间戳结构
type Timestamp struct {
	Sec  int64 // 精确秒级
	MSec int64 // 对齐100ms毫秒级
}

// GetTimestamp 获取当前对齐的时间戳
func GetTimestamp() Timestamp {
	return Timestamp{
		Sec:  GetSecond(),
		MSec: GetMillisecond(),
	}
}