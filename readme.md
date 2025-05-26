## 使用示例

### 注册加载器、创建器和存储器
### 绑定过期时间

```go
	loc, _ := time.LoadLocation("Asia/Shanghai")
    timestate.InitTimezoneTimer(loc)
    timestate.InitTimer(loc)
    nowtime := timestate.GetTimestamp()
    fmt.Println("nowtime" , nowtime)
    fmt.Println("nowtime" , timestate.GetNextDayTimestamp())
    fmt.Println("nowtime" , timestate.GetNextWeekTimestamp())
    fmt.Println("nowtime" , timestate.GetNextMonthTimestamp())
    userID := cycledata.UserID(1001)
    cycle := cycledata.DailyCycle
    typeKey := cycledata.TypeKey(1)

    now := time.Now()
    timestamp := int32(now.Unix())
    cycledata.RegisterLoader(cycledata.DailyCycle, cycledata.TypeKey(1), func(cycle cycledata.CycleType, typeKey cycledata.TypeKey, userID cycledata.UserID) *cycledata.PlayerData {
        // 模拟加载数据
        return &cycledata.PlayerData{
            UserID:     userID,
            UpdateTime: time.Now(),
            ExpireTime: timestamp + 24* 3600,
            MiscData:   make(map[string]interface{}),
        }
    })
    cycledata.RegisterCreator(cycledata.DailyCycle, cycledata.TypeKey(1), func(cycledata.UserID)(*cycledata.PlayerData){
        // 从数据库或缓存加载数据，示例返回空数据
        return &cycledata.PlayerData{
            UserID:     userID,
            MiscData:   make(map[string]interface{}),
            UpdateTime: time.Now(),
            ExpireTime: timestamp + 24* 3600,
        }
    })

    cycledata.RegisterStorer(cycledata.DailyCycle, cycledata.TypeKey(1), func(cycle cycledata.CycleType, typeKey cycledata.TypeKey, data *cycledata.PlayerData) error {
        // Store data to database or cache
        fmt.Printf("Storing %s data for user %d (Type: %d)\nDetails: %+v\n",
            cycle,
            data.UserID,
            typeKey,
            data.MiscData)

        return nil
    })
	cycledata.RegisterDefaultExpireFunc(cycledata.DailyCycle, cycledata.TypeKey(1), func() int32 {
		return int32(timestate.GetNextDayTimestamp())
	})
    // 获取数据（自动加载或创建）
    data := cycledata.GetData(cycle, typeKey, userID)
     if data == nil {
       fmt.Println("Appended achievement" , cycle, typeKey, userID , nil )
     }else{
       fmt.Println("Appended achievement" , cycle, typeKey, userID , data.MiscData )
     }
    // 更新数据字段
    cycledata.SetData(cycle, typeKey, userID , make(map[string]interface{}))

    // 条件追加 int32 切片示例
    ok := cycledata.AppendToInt32SliceIf(cycle, typeKey, userID, "achievements", 10, func(s []int32) bool {
    	return len(s) < 10
    })
    if ok {
    	fmt.Println("Appended achievement" , cycledata.GetDataValue(cycle, typeKey, userID))
    }

    // 刷新数据到存储器
    cycledata.Flush(cycle, typeKey)

    // 其他服务停止的时候调用一下
    cycledata.FlushAll()
```