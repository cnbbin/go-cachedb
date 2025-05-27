package statistic

import (
	"sync"
)

// StatisticHandler 统计处理器标识
type StatisticHandler int32

// StatisticType 统计类型标识
type StatisticType int32

// StatisticTypeCategory 统计类别标识
type StatisticTypeCategory int32

// handlerInfo 记录每个处理器相关的信息，包括静态类别和可选的worker函数
type handlerInfo struct {
	workerFunc func(statType StatisticType, categories []StatisticTypeCategory, addValue int32) []StatisticTypeCategory // 可选的worker处理函数，用于加工类别
	staticFunc func(statType StatisticType, categories []StatisticTypeCategory, addValue int32)                      // 可选的静态处理函数，用于自定义带参数处理	workerFunc func(statType StatisticType, categories []StatisticTypeCategory, addValue int32) []StatisticTypeCategory // 可选的worker处理函数，用于加工类别
	workerDoubleFunc func(statType StatisticType, categories []StatisticTypeCategory, addValue int32 , otherValue int32) // 可选的worker处理函数，用于加工类别
	staticDoubleFunc func(statType StatisticType, categories []StatisticTypeCategory, addValue int32 , otherValue int32) // 可选的静态处理函数，用于自定义带参数处理
	staticInfo map[StatisticType][]StatisticTypeCategory                                              // 静态缓存的类别信息，key为统计类型
}

// StatisticManager 管理多个处理器的类别注册和查询，支持静态缓存和动态查询
type StatisticManager struct {
	mu        sync.RWMutex
	registries map[StatisticHandler]*handlerInfo                                                    // 处理器注册信息
	queryFuncs map[StatisticHandler]func(statType StatisticType) []StatisticTypeCategory            // 可选的动态查询函数
}

/*
 * 全局周期处理器实例
 */
 var globalMgr *StatisticManager

func init() {
	globalMgr = NewStatisticManager()
}

// 全局单例管理器，方便统一管理和调用
func NewStatisticManager() *StatisticManager {
	return &StatisticManager{
		registries: make(map[StatisticHandler]*handlerInfo),
		queryFuncs: make(map[StatisticHandler]func(StatisticType) []StatisticTypeCategory),
	}
}