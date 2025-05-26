## 使用示例

### 注册加载器、创建器和存储器
### 绑定过期时间

```go
	loc, _ := time.LoadLocation("Asia/Shanghai")
    timestate.InitTimezoneTimer(loc)
    timestate.InitTimer(loc)
    nowtime := timestate.GetTimestamp()
    fmt.Println("nowtime" , nowtime)
    fmt.Println("nextdaytimestamp" , timestate.GetNextDayTimestamp())
    fmt.Println("nextweektimestamp" , timestate.GetNextWeekTimestamp())
    fmt.Println("nextmonthtimestamp" , timestate.GetNextMonthTimestamp())
    userID := cycledata.UserID(1001)
    cycle := cycledata.DailyCycle
    typeKey := cycledata.TypeKey(1)
    cycledata.RegisterLoader(cycledata.DailyCycle, cycledata.TypeKey(1), func(cycle cycledata.CycleType, typeKey cycledata.TypeKey, userID cycledata.UserID) *cycledata.PlayerData {
        // 模拟加载数据 （记得加载的时候判断对应过期时间和当前时间）
        return &cycledata.PlayerData{
            UserID:     userID,
            UpdateTime:  int32(timestate.GetSecond()),
            ExpireTime: int32(timestate.GetNextDayTimestamp()),
            MiscData:   make(map[string]interface{}),
        }
    })
    cycledata.RegisterCreator(cycledata.DailyCycle, cycledata.TypeKey(1), func(cycledata.UserID)(*cycledata.PlayerData){
        // 从数据库或缓存加载数据，示例返回空数据
        return &cycledata.PlayerData{
            UserID:     userID,
            MiscData:   make(map[string]interface{}),
            UpdateTime: int32(timestate.GetSecond()),
            ExpireTime: int32(timestate.GetNextDayTimestamp()),
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
    cycledata.RegisterCleanExpired(cycledata.DailyCycle, cycledata.TypeKey(1), func(cycle cycledata.CycleType, typeKey cycledata.TypeKey, data *cycledata.PlayerData) {
        // Store data to database or cache
        fmt.Printf("CleanExpiredr %s data for user %d (Type: %d)\nDetails: %+v\n",
            cycle,
            data.UserID,
            typeKey,
            data.MiscData)
    })

	cycledata.RegisterDefaultExpireFunc(cycledata.DailyCycle, cycledata.TypeKey(1), func() int32 {
		return int32(timestate.GetNextDayTimestamp())
	})

    timestate.RegisterDayCallback(func(t time.Time) {
        cycledata.CleanExpiredDataByType(cycledata.DailyCycle, cycledata.TypeKey(1))
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