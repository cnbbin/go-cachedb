package cycledata

import (
	"time"
	"fmt"
)

/*
// 只在新值大于旧值时更新
updated := UpdateIf(DailyCycle, 1, 12345, "score", 200, func(oldVal, newVal interface{}) bool {
	oldInt, _ := oldVal.(int)
	newInt, _ := newVal.(int)
	return newInt > oldInt
})

// 扣减金币（假设字段名为 "coins"）
ok := DecreaseIfEnoughInt(DailyCycle, 1, 12345, "coins", 50)

// 只有切片长度小于5才允许追加
ok := AppendToInt32SliceIf(DailyCycle, 1, 12345, "someInt32Slice", 42, func(s []int32) bool {
	// 只有切片长度小于5才允许追加
	return len(s) < 5
})
*/

func GameInfo(){
    // 注册加载器
    now := time.Now()
    timestamp := int32(now.Unix())
    RegisterLoader(DailyCycle, 1, func(cycle CycleType, typeKey TypeKey, userID UserID) *PlayerData {
        // 模拟加载数据
        return &PlayerData{
            UserID:     userID,
            UpdateTime: time.Now(),
            ExpireTime: timestamp + 24* 3600,
            MiscData:   make(map[string]interface{}),
        }
    })
    RegisterLoader(DailyCycle, 1, func(cycle CycleType, typeKey TypeKey, userID UserID) *PlayerData {
        // 从数据库或缓存加载数据，示例返回空数据
        return &PlayerData{
            UserID:     userID,
            MiscData:   make(map[string]interface{}),
            UpdateTime: time.Now(),
            ExpireTime: timestamp + 24* 3600,
        }
    })

    RegisterCreator(DailyCycle, 1, func(userID UserID) *PlayerData {
        // 创建新玩家数据
        return &PlayerData{
            UserID:     userID,
            MiscData:   make(map[string]interface{}),
            UpdateTime: time.Now(),
            ExpireTime:  timestamp + 24* 3600,
        }
    })

    RegisterStorer(func(cycle CycleType, typeKey TypeKey, data *PlayerData) error {
        // 存储数据到数据库或缓存
        fmt.Printf("Storing data for user %d, cycle %s, miscData %v , type %d\n", data.UserID, cycle, data.MiscData, typeKey)
        return nil
    })

    // 获取数据
    pd := GetData(DailyCycle, 1, 12345)
    pd.update("score", 100)
}