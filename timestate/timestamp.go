package timestate

import (
	"log"
	"sync/atomic"
	"time"
)

var (
	currentMsTimestamp  int64
	currentSecTimestamp int64
)

func InitMsSecTimer(tz *time.Location) {
	now := time.Now().In(tz)
	atomic.StoreInt64(&currentMsTimestamp, now.UnixMilli())
	atomic.StoreInt64(&currentSecTimestamp, now.Unix())

	go startMsTimer()
	go startSecTimer()
}

func startMsTimer() {
	log.Printf("[timestate] startMsTimer started")
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for now := range ticker.C {
		atomic.StoreInt64(&currentMsTimestamp, now.UnixMilli())
	}
}

func startSecTimer() {
	log.Printf("[timestate] startSecTimer started")
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for now := range ticker.C {
		atomic.StoreInt64(&currentSecTimestamp, now.Unix())
	}
}

func GetCurrentMs() int64 {
	return atomic.LoadInt64(&currentMsTimestamp)
}

func GetCurrentSec() int64 {
	return atomic.LoadInt64(&currentSecTimestamp)
}

type TimeSnapshot struct {
	Ms  int64
	Sec int64
}

func GetTimeSnapshot() TimeSnapshot {
	return TimeSnapshot{
		Ms:  GetCurrentMs(),
		Sec: GetCurrentSec(),
	}
}
