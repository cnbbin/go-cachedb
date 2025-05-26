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

    RegisterStorer(DailyCycle, 1, func(cycle CycleType, typeKey TypeKey, data *PlayerData) error {
        // Store data to database or cache
        fmt.Printf("Storing %s data for user %d (Type: %d)\nDetails: %+v\n", 
            cycle, 
            data.UserID, 
            typeKey,
            data.MiscData)
        
        // Actual storage implementation would go here
        // For example:
        // err := db.SavePlayerData(cycle, typeKey, data)
        // if err != nil {
        //     return fmt.Errorf("failed to save player data: %w", err)
        // }
        
        return nil
    })

    // 获取数据
    pd := GetData(DailyCycle, 1, 12345)
    pd.update("score", 100)
}