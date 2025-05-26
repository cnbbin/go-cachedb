/*
 * 获取玩家数据
 * 若玩家数据不存在，将通过已注册的创建器函数进行初始化加载或构造
 */
package cycledata

import (
    "golang.org/x/exp/maps"
)




 /*
  * GetData 获取指定周期、类型和玩家ID对应的数据
  * 返回值：
  *   - interface{} 表示存储的数据（需自行类型断言）
  *   - bool 是否存在该数据
  */

func GetData(cycle CycleType, typeKey TypeKey, userID UserID) *PlayerData {
	return globalHandler.
		getService(cycle, DefaultExpireFor(cycle , typeKey)).
		getCollection(typeKey).
		get(cycle, typeKey, userID)
}

func GetDataValue(cycle CycleType, typeKey TypeKey, userID UserID) (map[string]interface{}) {
	pb :=  globalHandler.
		getService(cycle, DefaultExpireFor(cycle , typeKey)).
		getCollection(typeKey).
		get(cycle, typeKey, userID)
	if pb == nil || pb.MiscData == nil {
		return make(map[string]interface{}) // return empty map if nil
	}
    return maps.Clone(pb.MiscData)

}
