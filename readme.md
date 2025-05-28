# 项目描述

```System Architecture Requirements
    系统架构需求
statistic (统计)  
-Statistic Layer (原子统计层)
--职责：高性能原子计数器，无状态轻量级操作
cache   (频繁数据) 作用: 定时落地，用于存储数据上报  
-Cache Layer (热数据缓存层)
--职责：临时聚合统计结果，定时刷盘
cycledata  (周期性/永久数据缓存)  
-CycleData Layer (周期数据层)
--职责：生命周期管理+注册模式核心
timestate  
-TimeState Layer (时间中枢)
--职责：统一时间管理和周期事件触发
```

## 使用示例

### 时间状态机 timestate
```go
    timestate.GetNextDayTimestamp()
    timestate.GetNextWeekTimestamp()
```

### 周期数据 cycledata
### 注册加载器、创建器和存储器
### 绑定过期时间状态机函数

```go
	loc, _ := time.LoadLocation("Asia/Shanghai")
    timestate.InitTimezoneTimer(loc)
    timestate.InitTimer(loc)
    nowtime := timestate.GetTimestamp()
    fmt.Println("nowtime" , nowtime)
    fmt.Println("nextdaytimestamp" , timestate.GetNextDayTimestamp())
    fmt.Println("nextweektimestamp" , timestate.GetNextWeekTimestamp())
    fmt.Println("nextmonthtimestamp" , timestate.GetNextMonthTimestamp())

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
        // CleanExpired
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
    userID := cycledata.UserID(1001)
    cycle := cycledata.DailyCycle
    typeKey := cycledata.TypeKey(1)
    // 获取数据（自动加载或创建）  会先调用loader 没数据后会调用Creator

    data := cycledata.GetData(cycle, typeKey, userID)
     if data == nil {
       fmt.Println("Appended achievement" , cycle, typeKey, userID , nil )
     }else{
       fmt.Println("Appended achievement" , cycle, typeKey, userID , data.MiscData )
     }

    // 直接设置数据
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

### 缓存数据 cache
### 注册 设置落地函数  初始化函数(上次数据太多兼容临时落地为json文件)

```go
    func InitData(){
        log.Printf("Initialized KV cache service")
    }
    // 调整为列表记录
    type ListHandler struct{}

    func (h *ListHandler) Flush(data []interface{}) error {

        // 这里实现你的刷新逻辑，比如将数据写入数据库、发送到远程服务等
        return nil
    }


    cache.RegisterListService("listPlayerCurrency", &ListHandler{}, 5*time.Second,  func(){InitData()})
```


## 使用示例
### 统计数据 statistic
```go

	const (
		PlayerLoginHandler statistic.StatisticHandler = 1001
		DailyLoginType     statistic.StatisticType    = 1
	)
	// 注册静态类别
	statistic.RegisterCategories(PlayerLoginHandler, DailyLoginType, []statistic.StatisticTypeCategory{
		1, 2, 3, // 可表示不同用户等级、渠道、区服等
	})
	// 注册 workerFunc：用于动态加工关联StatisticTypeCategory
	statistic.RegisterWorkerFunc(PlayerLoginHandler, func(t statistic.StatisticType, cats []statistic.StatisticTypeCategory , addValue int32) []statistic.StatisticTypeCategory {
		// 示例：给每个类别 +1000
		for _, c := range cats {
			fmt.Println("执行 RegisterWorkerFunc 动态处理 statistic.StatisticTypeCategory: %d" , c)
		}
		return cats
	})
	// 注册 queryFunc：用于回补类别（当未缓存时）
	statistic.RegisterQueryFunc(PlayerLoginHandler, func(t statistic.StatisticType) []statistic.StatisticTypeCategory {
		fmt.Println("执行 queryFunc 动态补充类别")
		return []statistic.StatisticTypeCategory{9, 10}
	})
	// 注册 staticFunc：用于执行统计行为
	statistic.RegisterStaticFunc(PlayerLoginHandler, func(t statistic.StatisticType, cats []statistic.StatisticTypeCategory, add int32) {
		fmt.Printf("统计行为触发 type=%v, addValue=%d, categories=%v\n", t, add, cats)
	})
	// 实际业务中调用统计逻辑
	addValue := int32(5)
	statistic.ApplyStaticFunc(PlayerLoginHandler, DailyLoginType, addValue)
	// 获取最终的类别结果（包含加工）
	finalCategories := statistic.GetCategories(PlayerLoginHandler, DailyLoginType)
	fmt.Printf("最终获取到的类别: %v\n", finalCategories)
```