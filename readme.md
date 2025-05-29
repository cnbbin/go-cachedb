# go-cachedb

**go-cachedb** 是一个基于 Go 语言开发的多层级缓存系统，旨在提供高性能的数据缓存解决方案。

## 项目简介

go-cachedb 采用模块化设计，主要包括以下四个核心组件：

- **Statistic Layer（原子统计层）**：提供高性能的原子计数器，用于无状态的轻量级操作。
- **Cache Layer（热数据缓存层）**：用于临时聚合统计结果，并定时将数据持久化。
- **CycleData Layer（周期数据层）**：管理周期性或永久性的数据缓存，支持生命周期管理和注册模式。
- **TimeState Layer（时间中枢）**：统一管理时间相关的操作，如周期事件的触发。

## 安装

使用 `go get` 命令安装：

```bash
go get github.com/cnbbin/go-cachedb
```

## 快速开始

### 初始化时间状态机

```go
loc, _ := time.LoadLocation("Asia/Shanghai")
timestate.InitTimezoneTimer(loc)
timestate.InitTimer(loc)

now := timestate.GetTimestamp()
fmt.Println("当前时间戳:", now)
fmt.Println("下一天时间戳:", timestate.GetNextDayTimestamp())
fmt.Println("下一周时间戳:", timestate.GetNextWeekTimestamp())
fmt.Println("下一月时间戳:", timestate.GetNextMonthTimestamp())
```

### 使用周期数据层

```go

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

## 贡献指南

欢迎对 go-cachedb 项目提出建议或贡献代码。请遵循以下步骤：

1. Fork 本仓库。
2. 创建一个新的分支：`git checkout -b feature/your-feature-name`。
3. 提交您的更改：`git commit -m 'Add some feature'`。
4. 推送到分支：`git push origin feature/your-feature-name`。
5. 提交 Pull Request。

## 许可证

本项目基于 MIT 许可证，详情请参阅 [LICENSE](https://github.com/cnbbin/go-cachedb/blob/main/LICENSE) 文件。

---

如需更详细的文档或示例，请访问项目的 [GitHub 页面](https://github.com/cnbbin/go-cachedb)。
