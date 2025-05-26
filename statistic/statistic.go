package statistic

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// ========== 核心数据结构 ==========

// StatType 统计类型标识
type StatType string

// Dimension 统计维度配置
type Dimension struct {
	Type       StatType
	TimeFormat string           // 时间格式
	KeyFunc    func(interface{}) []StatKey // 从数据记录提取统计键
}

// StatKey 统计键(复合主键)
type StatKey struct {
	Dimensions map[string]interface{} // 维度键值对
	TimeKey    string                // 时间维度
	StatType   StatType              // 统计类型
}

// StatValue 统计值
type StatValue struct {
	Metrics map[string]*atomic.Int64 // 指标集合
}

// StatCollector 统计收集器
type StatCollector struct {
	dimensions map[StatType]*Dimension // 注册的统计维度
	stats      sync.Map                // 存储统计数据
	timeCache  *TimeFormatCache        // 时间格式化缓存
}

// ========== 初始化方法 ==========

func NewStatCollector() *StatCollector {
	return &StatCollector{
		dimensions: make(map[StatType]*Dimension),
		timeCache:  NewTimeFormatCache(),
	}
}

// ========== 核心API ==========

// RegisterDimension 注册统计维度
func (s *StatCollector) RegisterDimension(d *Dimension) {
	s.dimensions[d.Type] = d
}

// Record 记录数据
func (s *StatCollector) Record(data interface{}) {
	for _, dim := range s.dimensions {
		keys := dim.KeyFunc(data)
		for _, key := range keys {
			s.update(key)
		}
	}
}

// GetStats 获取统计结果
func (s *StatCollector) GetStats(filter func(StatKey) bool) map[StatKey]StatValue {
	result := make(map[StatKey]StatValue)
	s.stats.Range(func(k, v interface{}) bool {
		key := k.(StatKey)
		if filter == nil || filter(key) {
			val := v.(*StatValue)
			// 深拷贝当前值
			copy := StatValue{Metrics: make(map[string]*atomic.Int64)}
			for m, v := range val.Metrics {
				copy.Metrics[m] = &atomic.Int64{}
				copy.Metrics[m].Store(v.Load())
			}
			result[key] = copy
		}
		return true
	})
	return result
}

// ========== 内部方法 ==========

func (s *StatCollector) update(key StatKey) {
	value, _ := s.stats.LoadOrStore(key, &StatValue{
		Metrics: make(map[string]*atomic.Int64),
	})
	stat := value.(*StatValue)
	
	// 这里简化处理，实际应根据业务更新具体指标
	if _, exists := stat.Metrics["count"]; !exists {
		stat.Metrics["count"] = &atomic.Int64{}
	}
	stat.Metrics["count"].Add(1)
}

// ========== 辅助结构 ==========

// TimeFormatCache 时间格式化缓存
type TimeFormatCache struct {
	cache sync.Map
}

func NewTimeFormatCache() *TimeFormatCache {
	return &TimeFormatCache{}
}

func (c *TimeFormatCache) Format(t time.Time, layout string) string {
	if v, ok := c.cache.Load(layout); ok {
		return v.(*timeFormat).Format(t)
	}
	
	f := &timeFormat{layout: layout}
	c.cache.Store(layout, f)
	return f.Format(t)
}

type timeFormat struct {
	layout string
	pool   sync.Pool
}

func (f *timeFormat) Format(t time.Time) string {
	if v := f.pool.Get(); v != nil {
		buf := v.(*[]byte)
		defer f.pool.Put(buf)
		*buf = t.AppendFormat((*buf)[:0], f.layout)
		return string(*buf)
	}
	
	buf := make([]byte, 0, len(f.layout)+10)
	f.pool.Put(&buf)
	return t.Format(f.layout)
}