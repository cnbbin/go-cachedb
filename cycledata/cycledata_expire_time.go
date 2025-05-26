/*
 * 默认过期时间注册器模块（Default Expire Time Registry）
 *
 * 模块用途：
 *   用于为不同的周期类型（CycleType）与类型键（TypeKey）组合注册默认过期时间函数，
 *   并在需要时获取对应的过期时间（单位为秒，int32）。
 *
 * 使用场景：
 *   - 数据缓存模块中，不同业务类型的数据需要配置不同的过期时长。
 *   - 支持每日、每周、每月、限时等周期类型。
 *   - 可根据具体类型键（如 "default"、"vip"）注册差异化策略。
 *
 * 主要接口：
 *   - RegisterDefaultExpireFunc(cycle, key, fn): 注册默认过期时间函数
 *   - DefaultExpireFor(cycle, key): 获取指定周期与类型键的默认过期时间（秒）
 *
 * 特性：
 *   - 线程安全（内部使用 sync.RWMutex 加锁）
 *   - 未注册返回 0，调用方可自行处理 fallback 逻辑
 *   - 默认过期时间函数支持自定义逻辑（如从配置读取、动态计算）
 *
 * 示例：
 *   RegisterDefaultExpireFunc(DailyCycle, "default", func() int32 { return 86400 }) // 注册每日过期为 86400 秒
 *   expire := DefaultExpireFor(DailyCycle, "default") // 获取每日的默认过期时间
 */
package cycledata

import (
	"sync"
)

// expireFunc 定义返回秒数的过期时间函数
type expireFunc func() int32

var (
	// expireFuncRegistry 存储每个周期类型和类型键对应的过期时间函数
	expireFuncRegistry = make(map[CycleType]map[TypeKey]expireFunc)

	// registryMu 保护注册表的并发访问
	registryMu sync.RWMutex
)

/*
 * RegisterDefaultExpireFunc 注册默认过期时间函数（单位：秒）
 * 用于为特定周期类型（CycleType）和类型键（TypeKey）注册一个返回过期时间的函数
 * 适用于不同周期和数据类型设置不同的默认过期策略
 *
 * 参数：
 *   - cycle: 周期类型（如每日、每周等）
 *   - key: 类型键（可用于区分具体类型，如 "default", "vip"）
 *   - fn: 返回默认过期时间（单位秒）的函数（如 func() int32 { return 86400 }）
 */
func RegisterDefaultExpireFunc(cycle CycleType, key TypeKey, fn func() int32) {
	registryMu.Lock()
	defer registryMu.Unlock()

	if _, ok := expireFuncRegistry[cycle]; !ok {
		expireFuncRegistry[cycle] = make(map[TypeKey]expireFunc)
	}
	expireFuncRegistry[cycle][key] = fn
}


/*
 * DefaultExpireFor 获取指定周期和类型键对应的默认过期时间（单位：秒）
 * 如果未注册过期时间函数，返回 0
 *
 * 参数：
 *   - cycle: 周期类型（如每日、每周等）
 *   - key: 类型键（如 "default"、"vip" 等）
 * 返回值：
 *   - int32: 默认过期时间（单位秒），未注册则为 0
 */
func DefaultExpireFor(cycle CycleType, key TypeKey) int32 {
	registryMu.RLock()
	defer registryMu.RUnlock()

	if typeMap, ok := expireFuncRegistry[cycle]; ok {
		if fn, ok := typeMap[key]; ok && fn != nil {
			return fn()
		}
	}
	return 0
}