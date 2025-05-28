package cycledata

import (
	"time"
)

/*
 * AppendToInt32SliceIf 尝试向指定 []int32 类型的键值添加元素 val
 * 条件：
 *   - val 不存在于切片中
 *   - cond(slice) 返回 true
 * 添加成功后更新时间
 * 返回是否添加成功
 */
func AppendToInt32SliceIf(cycle CycleType, typeKey TypeKey, userID UserID, key string, val int32, cond func([]int32) bool) bool {
	pd := GetData(cycle, typeKey, userID)
	if pd == nil {
		return false
	}

	pd.mu.Lock()
	defer pd.mu.Unlock()

	raw, ok := pd.MiscData[key]
	var slice []int32

	if ok {
		slice, ok = raw.([]int32)
		if !ok {
			// 字段存在，但不是[]int32类型，直接返回false
			return false
		}
	} else {
		// 字段不存在，初始化一个空切片
		slice = []int32{}
	}

	if !cond(slice) {
		return false
	}

	// 满足条件，追加值
	slice = append(slice, val)
	pd.MiscData[key] = slice
	pd.UpdateTime = time.Now()
	return true
}

/*
 * RemoveFromInt32SliceIf 尝试从指定 []int32 类型的键值中删除元素 val
 * 条件：
 *   - val 存在于切片中
 *   - cond(slice) 返回 true
 * 删除成功后更新时间
 * 返回是否删除成功
 */
func RemoveFromInt32SliceIf(cycle CycleType, typeKey TypeKey, userID UserID, key string, val int32, cond func([]int32) bool) bool {
	pd := GetData(cycle, typeKey, userID)
	if pd == nil {
		return false
	}

	pd.mu.Lock()
	defer pd.mu.Unlock()

	raw, ok := pd.MiscData[key]
	if !ok {
		return false
	}

	slice, ok := raw.([]int32)
	if !ok {
		// 字段存在，但不是[]int32类型，直接返回false
		return false
	}

	if !cond(slice) {
		return false
	}

	// 删除元素
	newSlice := make([]int32, 0, len(slice)-1)
	for _, v := range slice {
		if v != val {
			newSlice = append(newSlice, v)
		}
	}

	pd.MiscData[key] = newSlice
	pd.UpdateTime = time.Now()
	return true
}
