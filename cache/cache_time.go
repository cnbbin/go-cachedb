package cache

import (
	"sync"
	"sync/atomic"
	"time"
)

// TimeFormatCache 提供高效的时间格式化缓存
type TimeFormatCache struct {
	dateCache   sync.Map // map[int64]string (timestamp -> formatted date)
	hourCache   sync.Map // map[int64]string (timestamp -> formatted hour)
	cleanTicker *time.Ticker
	stopChan    chan struct{}
	expireAfter time.Duration
	lastClean   atomic.Int64
}

// NewTimeFormatCache 创建新的时间格式化缓存
// cleanInterval: 清理间隔
// expireAfter: 缓存过期时间
func NewTimeFormatCache(cleanInterval, expireAfter time.Duration) *TimeFormatCache {
	c := &TimeFormatCache{
		cleanTicker: time.NewTicker(cleanInterval),
		stopChan:    make(chan struct{}),
		expireAfter: expireAfter,
	}
	go c.cleanupWorker()
	return c
}

// FormatDate 格式化日期为 "20060102" 格式
func (c *TimeFormatCache) FormatDate(t time.Time) string {
	ts := t.Unix()
	if val, ok := c.dateCache.Load(ts); ok {
		return val.(string)
	}
	formatted := t.Format("20060102")
	c.dateCache.Store(ts, formatted)
	return formatted
}

// FormatHour 格式化时间为 "2006010215" 格式
func (c *TimeFormatCache) FormatHour(t time.Time) string {
	ts := t.Unix()
	if val, ok := c.hourCache.Load(ts); ok {
		return val.(string)
	}
	formatted := t.Format("2006010215")
	c.hourCache.Store(ts, formatted)
	return formatted
}

func (c *TimeFormatCache) cleanupWorker() {
	for {
		select {
		case <-c.cleanTicker.C:
			c.cleanup()
		case <-c.stopChan:
			c.cleanTicker.Stop()
			return
		}
	}
}

func (c *TimeFormatCache) cleanup() {
	expireTime := time.Now().Add(-c.expireAfter).Unix()

	c.dateCache.Range(func(key, value interface{}) bool {
		if key.(int64) < expireTime {
			c.dateCache.Delete(key)
		}
		return true
	})

	c.hourCache.Range(func(key, value interface{}) bool {
		if key.(int64) < expireTime {
			c.hourCache.Delete(key)
		}
		return true
	})

	c.lastClean.Store(time.Now().Unix())
}

// Stop 停止缓存清理goroutine
func (c *TimeFormatCache) Stop() {
	close(c.stopChan)
}
