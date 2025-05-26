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
 * RegisterStorer
 * 注册数据存储器
 */
func RegisterStorer(storer func(cycle CycleType, typeKey TypeKey, data *PlayerData) error) {
	storeData = storer
}
