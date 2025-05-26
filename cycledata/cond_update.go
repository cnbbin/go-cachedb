package cycledata

import (
	"time"
)


/*
 * UpdateIf 尝试根据条件更新指定周期、类型和玩家ID对应的数据
 * 仅当 cond 函数返回 true 时才执行更新
 * 参数：
 *   - newVal: 新数据
 *   - cond: 判断是否更新的函数，传入旧值和新值
 * 返回值：
 *   - bool 是否执行了更新
 */
func UpdateIf(cycle CycleType, typeKey TypeKey, userID UserID, key string, newVal interface{}, cond func(oldVal, newVal interface{}) bool) bool {
	pd := GetData(cycle, typeKey, userID)
	if pd == nil {
		return false
	}

	pd.mu.Lock()
	defer pd.mu.Unlock()

	oldVal, ok := pd.MiscData[key]
	if !ok {
		oldVal = nil
	}

	if !cond(oldVal, newVal) {
		return false
	}

	pd.MiscData[key] = newVal
	pd.UpdateTime = time.Now()
	return true
}
