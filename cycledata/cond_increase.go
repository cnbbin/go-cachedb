package cycledata

import (
	"time"
)

/*
 * IncreaseIfCondInt 尝试增加 int 类型数值
 * 仅当 cond(current) 返回 true 时才执行增加操作
 *
 * 参数：
 *   - cycle: 周期类型（如每日、每周等）
 *   - typeKey: 类型标识（如积分、次数等）
 *   - userID: 用户 ID
 *   - key: 具体数据的键
 *   - amount: 增加的数值
 *   - cond: 条件函数，传入当前值，返回是否允许增加
 *
 * 返回值：
 *   - bool: 是否成功增加
 */
func IncreaseIfCondInt(cycle CycleType, typeKey TypeKey, userID UserID, key string, amount int, cond func(current int) bool) bool {
	pd := GetData(cycle, typeKey, userID)
	if pd == nil {
		return false
	}

	pd.mu.Lock()
	defer pd.mu.Unlock()

	rawVal, ok := pd.MiscData[key]
	oldVal := 0
	if ok {
		oldVal, ok = rawVal.(int)
		if !ok {
			return false
		}
	}

	if !cond(oldVal) {
		return false
	}

	pd.MiscData[key] = oldVal + amount
	pd.UpdateTime = time.Now()
	return true
}

/*
 * IncreaseIfCondInt32 尝试增加 int32 类型数值
 * 仅当 cond(current) 返回 true 时才执行增加操作
 *
 * 参数和逻辑同 IncreaseIfCondInt
 */
func IncreaseIfCondInt32(cycle CycleType, typeKey TypeKey, userID UserID, key string, amount int32, cond func(current int32) bool) bool {
	pd := GetData(cycle, typeKey, userID)
	if pd == nil {
		return false
	}

	pd.mu.Lock()
	defer pd.mu.Unlock()

	rawVal, ok := pd.MiscData[key]
	oldVal := int32(0)
	if ok {
		oldVal, ok = rawVal.(int32)
		if !ok {
			return false
		}
	}

	if !cond(oldVal) {
		return false
	}

	pd.MiscData[key] = oldVal + amount
	pd.UpdateTime = time.Now()
	return true
}

/*
 * IncreaseIfCondInt64 尝试增加 int64 类型数值
 * 仅当 cond(current) 返回 true 时才执行增加操作
 *
 * 参数和逻辑同 IncreaseIfCondInt
 */
func IncreaseIfCondInt64(cycle CycleType, typeKey TypeKey, userID UserID, key string, amount int64, cond func(current int64) bool) bool {
	pd := GetData(cycle, typeKey, userID)
	if pd == nil {
		return false
	}

	pd.mu.Lock()
	defer pd.mu.Unlock()

	rawVal, ok := pd.MiscData[key]
	oldVal := int64(0)
	if ok {
		oldVal, ok = rawVal.(int64)
		if !ok {
			return false
		}
	}

	if !cond(oldVal) {
		return false
	}

	pd.MiscData[key] = oldVal + amount
	pd.UpdateTime = time.Now()
	return true
}

/*
 * IncreaseIfCondFloat64 尝试增加 float64 类型数值
 * 仅当 cond(current) 返回 true 时才执行增加操作
 *
 * 参数和逻辑同 IncreaseIfCondInt
 */
func IncreaseIfCondFloat64(cycle CycleType, typeKey TypeKey, userID UserID, key string, amount float64, cond func(current float64) bool) bool {
	pd := GetData(cycle, typeKey, userID)
	if pd == nil {
		return false
	}

	pd.mu.Lock()
	defer pd.mu.Unlock()

	rawVal, ok := pd.MiscData[key]
	oldVal := float64(0)
	if ok {
		oldVal, ok = rawVal.(float64)
		if !ok {
			return false
		}
	}

	if !cond(oldVal) {
		return false
	}

	pd.MiscData[key] = oldVal + amount
	pd.UpdateTime = time.Now()
	return true
}
