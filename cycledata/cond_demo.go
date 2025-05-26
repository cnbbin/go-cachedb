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
    RegisterLoader(DailyCycle, TypeKey(1), func(cycle CycleType, typeKey TypeKey, userID UserID) *PlayerData {
        // 模拟加载数据
        return &PlayerData{
            UserID:     userID,
            UpdateTime: time.Now(),
            ExpireTime: timestamp + 24* 3600,
            MiscData:   make(map[string]interface{}),
        }
    })
    RegisterCreator(DailyCycle, TypeKey(1), func(UserID)(*PlayerData){
        // 从数据库或缓存加载数据，示例返回空数据
        return &PlayerData{
            UserID:     userID,
            MiscData:   make(map[string]interface{}),
            UpdateTime: time.Now(),
            ExpireTime: timestamp + 24* 3600,
        }
    })

    RegisterStorer(DailyCycle, TypeKey(1), func(cycle CycleType, typeKey TypeKey, data *PlayerData) error {
        // Store data to database or cache
        fmt.Printf("Storing %s data for user %d (Type: %d)\nDetails: %+v\n",
            cycle,
            data.UserID,
            typeKey,
            data.MiscData)

        return nil
    })
    userID := cycledata.UserID(1001)
    cycle := cycledata.DailyCycle
    typeKey := cycledata.TypeKey(1)
    // 获取数据（自动加载或创建）
    data := GetData(cycle, typeKey, userID)
     if data == nil {
            fmt.Println("Appended achievement" , cycle, typeKey, userID , nil )
     }
    // 更新数据字段
    SetData(cycle, typeKey, userID , make(map[string]interface{}))

    // 条件追加 int32 切片示例
    ok := AppendToInt32SliceIf(cycle, typeKey, userID, "achievements", 10, func(s []int32) bool {
    	return len(s) < 10
    })
    if ok {
    	fmt.Println("Appended achievement" , cycledata.GetDataValue(cycle, typeKey, userID))
    }

    // 刷新数据到存储器
    cycledata.Flush(cycle, typeKey)

    // 其他服务停止的时候调用一下
    cycledata.FlushAll()
}