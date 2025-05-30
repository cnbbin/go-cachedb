/*
 * 覆盖玩家数据（使用注册创建器，并注入 MiscData）
 */
package cycledata

import "time"

func SetData(cycle CycleType, typeKey TypeKey, userID UserID, miscData map[string]interface{}) bool {
	return globalHandler.
		getService(cycle, DefaultExpireFor(cycle, typeKey)).
		getCollection(typeKey).
		set(cycle, typeKey, userID, miscData)
}

/*
 * SetWithAllMiscData 尝试向指定 map[int32]int32 类型的键值设置元素 key:val
 * 自定义条件：
 *   - 返回是否设置成功
 *   - 新的newMap
 *   - 是否改成时间
 */
func SetWithAllMiscData(cycle CycleType, typeKey TypeKey, userID UserID,
	cond func(time.Time, map[string]interface{}) (bool, map[string]interface{}, bool)) bool {

	// 获取数据指针
	pd := GetData(cycle, typeKey, userID)
	if pd == nil {
		return false
	}

	pd.mu.Lock()
	defer pd.mu.Unlock()

	// 直接使用原始数据，但通过闭包限制修改
	success, newMiscData, changeTimeBool := cond(pd.UpdateTime, pd.MiscData)
	if !success {
		return false
	}

	// 检查新数据是否真的被修改了
	if newMiscData != nil && &newMiscData != &pd.MiscData {
		pd.MiscData = newMiscData
	}

	if changeTimeBool {
		pd.UpdateTime = time.Now()
	}
	return true
}
