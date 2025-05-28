package cycledata

import (
	"time"
)

/*
 * SetInInt32MapIf 尝试向指定 map[int32]int32 类型的键值设置元素 key:val
 * 条件：
 *   - key 不存在于map中
 *   - cond(m) 返回 true
 * 设置成功后更新时间
 * 返回是否设置成功
 */
func SetInInt32MapIf(cycle CycleType, typeKey TypeKey, userID UserID, mapKey string, key, val int32, cond func(map[int32]int32) bool) bool {
	pd := GetData(cycle, typeKey, userID)
	if pd == nil {
		return false
	}

	pd.mu.Lock()
	defer pd.mu.Unlock()

	raw, ok := pd.MiscData[mapKey]
	var m map[int32]int32

	if ok {
		m, ok = raw.(map[int32]int32)
		if !ok {
			// 字段存在，但不是map[int32]int32类型，直接返回false
			return false
		}
	} else {
		// 字段不存在，初始化一个空map
		m = make(map[int32]int32)
	}

	if !cond(m) {
		return false
	}

	// 检查key是否已存在
	if _, exists := m[key]; exists {
		return false
	}

	// 满足条件，设置值
	m[key] = val
	pd.MiscData[mapKey] = m
	pd.UpdateTime = time.Now()
	return true
}

/*
 * SetWithCDInInt32MapIf 尝试向指定 map[int32]int32 类型的键值设置元素 key:val
 * 条件：
 *   - lastUpdateLimitSec 与上次cd间隔限制
 *   - key 不存在于map中
 *   - cond(m) 返回 true
 * 设置成功后更新时间
 * 返回是否设置成功
 */
func SetWithCDInInt32MapIf(cycle CycleType, typeKey TypeKey, userID UserID, mapKey string, key, val int32, lastUpdateLimitSec int, cond func(map[int32]int32) bool) bool {
	pd := GetData(cycle, typeKey, userID)
	if pd == nil {
		return false
	}

	pd.mu.Lock()
	defer pd.mu.Unlock()

	raw, ok := pd.MiscData[mapKey]
	var m map[int32]int32

	if ok {
		m, ok = raw.(map[int32]int32)
		if !ok {
			// 字段存在，但不是map[int32]int32类型，直接返回false
			return false
		}
	} else {
		// 字段不存在，初始化一个空map
		m = make(map[int32]int32)
	}

	// 检查时间间隔
	if lastUpdateLimitSec > 0 {
		if time.Since(pd.UpdateTime) > time.Duration(lastUpdateLimitSec)*time.Second {
			return false
		}
	}

	if !cond(m) {
		return false
	}

	// 检查key是否已存在
	if _, exists := m[key]; exists {
		return false
	}

	// 满足条件，设置值
	m[key] = val
	pd.MiscData[mapKey] = m
	pd.UpdateTime = time.Now()
	return true
}

/*
 * RemoveWithCDFromInt32MapIf 尝试从指定 map[int32]int32 类型的键值中删除元素 key
 * 条件：
 *   - lastUpdateLimitSec 与上次cd间隔限制
 *   - key 存在于map中
 *   - cond(m) 返回 true
 * 删除成功后更新时间
 * 返回是否删除成功
 */
func RemoveWithCDFromInt32MapIf(cycle CycleType, typeKey TypeKey, userID UserID, mapKey string, key int32, lastUpdateLimitSec int, cond func(map[int32]int32) bool) bool {
	pd := GetData(cycle, typeKey, userID)
	if pd == nil {
		return false
	}

	pd.mu.Lock()
	defer pd.mu.Unlock()

	raw, ok := pd.MiscData[mapKey]
	if !ok {
		return false
	}

	m, ok := raw.(map[int32]int32)
	if !ok {
		// 字段存在，但不是map[int32]int32类型，直接返回false
		return false
	}

	// 检查时间间隔
	if lastUpdateLimitSec > 0 {
		if time.Since(pd.UpdateTime) > time.Duration(lastUpdateLimitSec)*time.Second {
			return false
		}
	}

	if !cond(m) {
		return false
	}

	// 检查key是否存在
	if _, exists := m[key]; !exists {
		return false
	}

	// 删除元素
	delete(m, key)
	pd.MiscData[mapKey] = m
	pd.UpdateTime = time.Now()
	return true
}

/*
 * UpdateInInt32MapIf 尝试更新指定 map[int32]int32 类型中已存在的键值 key:val
 * 条件：
 *   - key 存在于map中
 *   - cond(m) 返回 true
 * 更新成功后更新时间
 * 返回是否更新成功
 */
func UpdateInInt32MapIf(cycle CycleType, typeKey TypeKey, userID UserID, mapKey string, key, val int32, cond func(map[int32]int32) bool) bool {
	pd := GetData(cycle, typeKey, userID)
	if pd == nil {
		return false
	}

	pd.mu.Lock()
	defer pd.mu.Unlock()

	raw, ok := pd.MiscData[mapKey]
	if !ok {
		return false
	}

	m, ok := raw.(map[int32]int32)
	if !ok {
		// 字段存在，但不是map[int32]int32类型，直接返回false
		return false
	}

	if !cond(m) {
		return false
	}

	// 检查key是否存在
	if _, exists := m[key]; !exists {
		return false
	}

	// 满足条件，更新值
	m[key] = val
	pd.MiscData[mapKey] = m
	pd.UpdateTime = time.Now()
	return true
}
