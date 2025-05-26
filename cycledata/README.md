# cycledata

`cycledata` 是一个用于管理用户周期性数据缓存与更新的 Go 包，支持每日、每周、每月等多周期维度的数据存储和自动清理。适合游戏、活动系统等需要周期性统计和数据管理的场景。

---

## 特性

- 支持多周期（每日、每周、每月）数据分类管理
- 按周期和数据类型分组管理玩家数据
- 支持自定义数据加载器、创建器和存储器注册
- 自动过期清理，支持定时异步保存数据
- 线程安全设计，支持高并发访问
- 支持扩展字段（`map[string]interface{}`）存储灵活数据
- 提供条件更新函数，支持复杂数据变更逻辑

---

## 安装

```bash
go get github.com/your-repo/cycledata
```

---

## 使用示例

### 注册加载器、创建器和存储器

```go
cycledata.RegisterLoader(cycledata.DailyCycle, 1, func(cycle cycledata.CycleType, typeKey cycledata.TypeKey, userID cycledata.UserID) *cycledata.PlayerData {
	// 从数据库或缓存加载数据或中心数据节点,读取数据,读取不到按照示例返回空数据
	return &cycledata.PlayerData{
		UserID:     userID,
		MiscData:   make(map[string]interface{}),
		UpdateTime: time.Now(),
		ExpireTime: time.Now().Add(24 * time.Hour),
	}
})

cycledata.RegisterCreator(cycledata.DailyCycle, 1, func(userID cycledata.UserID) *cycledata.PlayerData {
	// 从数据库或缓存加载数据或中心数据节点,创建新玩家数据 无法创建或者报错示例返回空数据
	return &cycledata.PlayerData{
		UserID:     userID,
		MiscData:   make(map[string]interface{}),
		UpdateTime: time.Now(),
		ExpireTime: time.Now().Add(24 * time.Hour),
	}
})

cycledata.RegisterStorer(func(cycle cycledata.CycleType, typeKey cycledata.TypeKey, data *cycledata.PlayerData) error {
	// 存储数据到数据库或缓存
	fmt.Printf("Storing data for user %d, cycle %s, type %d\n", data.UserID, cycle, typeKey)
	return nil
})
```

### 读取与更新数据

```go
userID := cycledata.UserID(1001)
cycle := cycledata.DailyCycle
typeKey := cycledata.TypeKey(1)

// 获取数据（自动加载或创建）
data := cycledata.GetData(cycle, typeKey, userID)

// 更新数据字段
data.update("score", 500)

// 条件追加 int32 切片示例
ok := cycledata.AppendToInt32SliceIf(cycle, typeKey, userID, "achievements", 10, func(s []int32) bool {
	return len(s) < 10
})
if ok {
	fmt.Println("Appended achievement")
}

// 刷新数据到存储器
cycledata.Flush(cycle, typeKey)

// 其他服务停止的时候调用一下
cycledata.FlushAll()
```

---

## API

- `RegisterLoader(cycle CycleType, typeKey TypeKey, loader func(cycle CycleType, typeKey TypeKey, userID UserID) *PlayerData)`
- `RegisterCreator(cycle CycleType, typeKey TypeKey, creator func(userID UserID) *PlayerData)`
- `RegisterStorer(storer func(cycle CycleType, typeKey TypeKey, data *PlayerData) error)`
- `GetData(cycle CycleType, typeKey TypeKey, userID UserID) *PlayerData`
- `Flush(cycle CycleType, typeKey TypeKey)`
- `FlushAll()`
- `AppendToInt32SliceIf(cycle CycleType, typeKey TypeKey, userID UserID, key string, val int32, cond func([]int32) bool) bool`
- `DecreaseIfEnoughInt(cycle CycleType, typeKey TypeKey, userID UserID, key string, amount int) bool`
- `IncreaseIfCondInt(cycle, typeKey, userID, key, int32(amount), func(int)) bool`
- `IncreaseIfCondInt32(cycle, typeKey, userID, key, int32(amount), func(int32)) bool`
- `IncreaseIfCondInt64(cycle, typeKey, userID, key, int32(amount), func(int64)) bool`
- `IncreaseIfCondFloat64(cycle, typeKey, userID, key, int32(amount), func(float64)) bool`
- `UpdateIf(cycle CycleType, typeKey TypeKey, userID UserID, key string, newVal interface{}, cond func(oldVal, newVal interface{}) bool) bool`
- `GetDataExist(cycle CycleType, typeKey TypeKey, userID UserID) (interface{}, bool)`

---

## 许可

MIT License
