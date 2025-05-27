## 使用示例

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