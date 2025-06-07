package cycledata

/*
 * RegisterLoader
 * 注册数据加载器
 */
func RegisterLoader(cycle CycleType, typeKey TypeKey, loader func(cycle CycleType, typeKey TypeKey, userID UserID) *PlayerData) {
	if _, ok := loaders[cycle]; !ok {
		loaders[cycle] = make(map[TypeKey]func(CycleType, TypeKey, UserID) *PlayerData)
	}
	loaders[cycle][typeKey] = loader
}

/*
 * RegisterCreator
 * 注册数据创建器
 */
func RegisterCreator(cycle CycleType, typeKey TypeKey, creator func(userID UserID) *PlayerData) {
	if _, ok := creators[cycle]; !ok {
		creators[cycle] = make(map[TypeKey]func(UserID) *PlayerData)
	}
	creators[cycle][typeKey] = creator
}

/*
 * RegisterStorer registers a storage function for a specific cycle type and data type
 *
 * Parameters:
 *   cycle - The cycle type (e.g., daily, weekly, monthly)
 *   typeKey - The data type identifier
 *   store - The storage function that handles persisting player data
 */
func RegisterStorer(cycle CycleType, typeKey TypeKey,
	store func(cycle CycleType, typeKey TypeKey, data *PlayerData) error) {

	// Initialize the inner map if this cycle type hasn't been registered before
	if _, ok := stores[cycle]; !ok {
		stores[cycle] = make(map[TypeKey]func(CycleType, TypeKey, *PlayerData) error)
	}

	// Register the store function for this specific cycle and type
	stores[cycle][typeKey] = store
}

/*
 * 注册自定义过期处理函数
 */
func RegisterCleanExpired(cycle CycleType, typeKey TypeKey, handler func(cycle CycleType, typeKey TypeKey, data *PlayerData)) {
	if _, ok := cleanExpireds[cycle]; !ok {
		cleanExpireds[cycle] = make(map[TypeKey]func(cycle CycleType, typeKey TypeKey, data *PlayerData))
	}
	cleanExpireds[cycle][typeKey] = handler
}
