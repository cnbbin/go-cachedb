/*
 * 覆盖玩家数据（使用注册创建器，并注入 MiscData）
 */
 package cycledata

import (

)

func SetData(cycle CycleType, typeKey TypeKey, userID UserID,  miscData map[string]interface{}) (bool){
	return globalHandler.
		getService(cycle, DefaultExpireFor(cycle , typeKey)).
		getCollection(typeKey).
		set(cycle, typeKey, userID , miscData)
}