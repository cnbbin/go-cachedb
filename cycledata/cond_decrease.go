package cycledata

import (
	"time"
)


/*
 * DecreaseIfEnoughInt 尝试减少 int 类型数值
 * 仅当当前值 >= amount 时才执行扣减
 * 返回值：
 *   - bool 是否扣减成功
 */
func DecreaseIfEnoughInt(cycle CycleType, typeKey TypeKey, userID UserID, key string, amount int) bool {
	pd := GetData(cycle, typeKey, userID)
	if pd == nil {
		return false
	}

	pd.mu.Lock()
	defer pd.mu.Unlock()

	rawVal, ok := pd.MiscData[key]
	if !ok {
		return false
	}

	// 尝试转换成 int
	oldVal, ok := rawVal.(int)
	if !ok {
		// 尝试支持 int64
		if val64, ok64 := rawVal.(int64); ok64 {
			oldVal = int(val64)
		} else {
			return false
		}
	}

	if oldVal < amount {
		return false
	}

	pd.MiscData[key] = oldVal - amount
	pd.UpdateTime = time.Now()
	return true
}

/*
 * DecreaseIfEnoughInt32 尝试减少 int32 类型数值
 * 仅当当前值 >= amount 时才执行扣减
 * 返回值：
 *   - bool 是否扣减成功
 */
func DecreaseIfEnoughInt32(cycle CycleType, typeKey TypeKey, userID UserID, key string, amount int32) bool {
	pd := GetData(cycle, typeKey, userID)
	if pd == nil {
		return false
	}

	pd.mu.Lock()
	defer pd.mu.Unlock()

	rawVal, ok := pd.MiscData[key]
	if !ok {
		return false
	}

	oldVal, ok := rawVal.(int32)
	if !ok {
		// 兼容 int 和 int64 转 int32
		switch v := rawVal.(type) {
		case int:
			oldVal = int32(v)
		case int64:
			oldVal = int32(v)
		default:
			return false
		}
	}

	if oldVal < amount {
		return false
	}

	pd.MiscData[key] = oldVal - amount
	pd.UpdateTime = time.Now()
	return true
}

/*
 * DecreaseIfEnoughFloat64 尝试减少 float64 类型数值
 * 仅当当前值 >= amount 时才执行扣减
 * 返回值：
 *   - bool 是否扣减成功
 */
func DecreaseIfEnoughFloat64(cycle CycleType, typeKey TypeKey, userID UserID, key string, amount float64) bool {
	pd := GetData(cycle, typeKey, userID)
	if pd == nil {
		return false
	}

	pd.mu.Lock()
	defer pd.mu.Unlock()

	rawVal, ok := pd.MiscData[key]
	if !ok {
		return false
	}

	oldVal, ok := rawVal.(float64)
	if !ok {
		// 兼容 float32 转 float64
		if val32, ok32 := rawVal.(float32); ok32 {
			oldVal = float64(val32)
		} else {
			return false
		}
	}

	if oldVal < amount {
		return false
	}

	pd.MiscData[key] = oldVal - amount
	pd.UpdateTime = time.Now()
	return true
}
