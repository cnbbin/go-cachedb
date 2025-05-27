package statistic


// RegisterCategories 为指定处理器和类型注册静态类别信息，支持追加
func (m *StatisticManager) RegisterCategories(handler StatisticHandler, t StatisticType, categories []StatisticTypeCategory) {
	m.mu.Lock()
	defer m.mu.Unlock()

	hInfo, exists := m.registries[handler]
	if !exists {
		// 若对应处理器尚未注册，初始化handlerInfo
		hInfo = &handlerInfo{
			staticInfo: make(map[StatisticType][]StatisticTypeCategory),
		}
		m.registries[handler] = hInfo
	}

	// 追加类别
	hInfo.staticInfo[t] = append(hInfo.staticInfo[t], categories...)
}


// RegisterCategories 为指定处理器和类型注册静态类别信息，重置
func (m *StatisticManager) ResetRegisterCategories(handler StatisticHandler, t StatisticType, categories []StatisticTypeCategory) {
	m.mu.Lock()
	defer m.mu.Unlock()

	hInfo, exists := m.registries[handler]
	if !exists {
		// 若对应处理器尚未注册，初始化handlerInfo
		hInfo = &handlerInfo{
			staticInfo: make(map[StatisticType][]StatisticTypeCategory),
		}
		m.registries[handler] = hInfo
	}

	// 追加类别
	hInfo.staticInfo[t] = categories
}

// RegisterQueryFunc 为指定处理器注册一个动态查询函数，当静态缓存未命中时调用
func (m *StatisticManager) RegisterQueryFunc(handler StatisticHandler, f func(statType StatisticType) []StatisticTypeCategory) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.queryFuncs[handler] = f
}

// RegisterWorkerFunc 为指定处理器注册worker函数，用于对静态类别数据进行二次加工处理
func (m *StatisticManager) RegisterWorkerFunc(handler StatisticHandler, f func(statType StatisticType, categories []StatisticTypeCategory, addValue int32) []StatisticTypeCategory) {
	m.mu.Lock()
	defer m.mu.Unlock()

	hInfo, exists := m.registries[handler]
	if !exists {
		hInfo = &handlerInfo{
			staticInfo: make(map[StatisticType][]StatisticTypeCategory),
		}
		m.registries[handler] = hInfo
	}
	hInfo.workerFunc = f
}

// RegisterStaticFunc 注册 staticFunc，用于带额外参数的处理逻辑
func (m *StatisticManager) RegisterStaticFunc(handler StatisticHandler, f func(statType StatisticType, categories []StatisticTypeCategory, addValue int32)) {
	m.mu.Lock()
	defer m.mu.Unlock()

	hInfo, exists := m.registries[handler]
	if !exists {
		hInfo = &handlerInfo{
			staticInfo: make(map[StatisticType][]StatisticTypeCategory),
		}
		m.registries[handler] = hInfo
	}
	hInfo.staticFunc = f
}


// ==== 全局包装函数，方便直接调用全局单例管理器 ====

// RegisterCategories 注册静态类别
func RegisterCategories(handler StatisticHandler, t StatisticType, categories []StatisticTypeCategory) {
	GetGlobalManager().RegisterCategories(handler, t, categories)
}

// RegisterCategories 注册静态类别
func ResetRegisterCategories(handler StatisticHandler, t StatisticType, categories []StatisticTypeCategory) {
	GetGlobalManager().ResetRegisterCategories(handler, t, categories)
}


// RegisterQueryFunc 注册动态查询函数
func RegisterQueryFunc(handler StatisticHandler, f func(statType StatisticType) []StatisticTypeCategory) {
	GetGlobalManager().RegisterQueryFunc(handler, f)
}

// RegisterWorkerFunc 注册worker函数
func RegisterWorkerFunc(handler StatisticHandler, f func(statType StatisticType, categories []StatisticTypeCategory, addValue int32) []StatisticTypeCategory) {
	GetGlobalManager().RegisterWorkerFunc(handler, f)
}

// GetCategories 获取类别
func GetCategories(handler StatisticHandler, t StatisticType) []StatisticTypeCategory {
	return GetGlobalManager().GetCategories(handler, t)
}

// RegisterStaticFunc 注册 staticFunc
func RegisterStaticFunc(handler StatisticHandler, f func(statType StatisticType, categories []StatisticTypeCategory, addValue int32)) {
	GetGlobalManager().RegisterStaticFunc(handler, f)
}