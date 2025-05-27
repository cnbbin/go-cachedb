package statistic



// GetGlobalManager 返回全局单例管理器实例
func GetGlobalManager() *StatisticManager {
	return globalMgr
}

// GetCategories 获取指定处理器和类型的类别列表，优先返回静态缓存，若无则调用动态查询函数，且支持workerFunc二次加工
func (m *StatisticManager) GetCategories(handler StatisticHandler, t StatisticType) []StatisticTypeCategory {
	m.mu.RLock()
	hInfo, exists := m.registries[handler]
	m.mu.RUnlock()

	if exists {
		// 先尝试获取静态缓存的类别
		if cats, ok := hInfo.staticInfo[t]; ok && len(cats) > 0 {
			return cats
		}
	}

	// 静态缓存没有，尝试调用动态查询函数获取类别
	m.mu.RLock()
	queryFunc, hasQueryFunc := m.queryFuncs[handler]
	m.mu.RUnlock()

	// 如果没有注册动态查询函数，返回nil
	if !hasQueryFunc || queryFunc == nil {
		return nil
	}

	cats := queryFunc(t)
	if len(cats) == 0 {
		return nil
	}

	// 缓存查询结果到静态缓存中
	m.mu.Lock()
	defer m.mu.Unlock()
	hInfo, exists = m.registries[handler]
	if !exists {
		hInfo = &handlerInfo{
			staticInfo: make(map[StatisticType][]StatisticTypeCategory),
		}
		m.registries[handler] = hInfo
	}

	if _, ok := hInfo.staticInfo[t]; !ok {
		hInfo.staticInfo[t] = cats
	}

	return cats
}




// ApplyStaticFunc 尝试调用指定 handler 的 staticFunc（若有 workerFunc 先加工）
func (m *StatisticManager) ApplyStaticFunc(handler StatisticHandler, t StatisticType, addValue int32) {
	// Step 1: 尝试从 registries 获取 handlerInfo
	m.mu.RLock()
	hInfo, exists := m.registries[handler]
	m.mu.RUnlock()

	// Step 2: 若不存在则 fallback 到 queryFunc 获取类别并注册
	if !exists {
		m.mu.RLock()
		queryFunc, hasQueryFunc := m.queryFuncs[handler]
		m.mu.RUnlock()

		if !hasQueryFunc || queryFunc == nil {
			return
		}

		categories := queryFunc(t)
		if len(categories) == 0 {
			return
		}

		m.mu.Lock()
		hInfo, exists = m.registries[handler]
		if !exists {
			hInfo = &handlerInfo{
				staticInfo: make(map[StatisticType][]StatisticTypeCategory),
			}
			m.registries[handler] = hInfo
		}
		if _, ok := hInfo.staticInfo[t]; !ok {
			hInfo.staticInfo[t] = categories
		}
		m.mu.Unlock()
	}

	// Step 3: 重新获取 handlerInfo 和处理函数
	m.mu.RLock()
	hInfo = m.registries[handler]
	categories := hInfo.staticInfo[t]
	workerFunc := hInfo.workerFunc
	staticFunc := hInfo.staticFunc
	m.mu.RUnlock()

	if len(categories) == 0 || staticFunc == nil {
		return
	}

	// Step 4: 若有 workerFunc，则先处理
	if workerFunc != nil {
		categories = workerFunc(t, categories, addValue)
	}

	// Step 5: 调用 staticFunc
	staticFunc(t, categories, addValue)
}



// ApplyStaticFunc 尝试调用指定 handler 的 staticFunc（若有 workerFunc 先加工）
func (m *StatisticManager) ApplyStaticDoubleFunc(handler StatisticHandler, t StatisticType, addValue int32, otherValue int32) {
	// Step 1: 尝试从 registries 获取 handlerInfo
	m.mu.RLock()
	hInfo, exists := m.registries[handler]
	m.mu.RUnlock()

	// Step 2: 若不存在则 fallback 到 queryFunc 获取类别并注册
	if !exists {
		m.mu.RLock()
		queryFunc, hasQueryFunc := m.queryFuncs[handler]
		m.mu.RUnlock()

		if !hasQueryFunc || queryFunc == nil {
			return
		}

		categories := queryFunc(t)
		if len(categories) == 0 {
			return
		}

		m.mu.Lock()
		hInfo, exists = m.registries[handler]
		if !exists {
			hInfo = &handlerInfo{
				staticInfo: make(map[StatisticType][]StatisticTypeCategory),
			}
			m.registries[handler] = hInfo
		}
		if _, ok := hInfo.staticInfo[t]; !ok {
			hInfo.staticInfo[t] = categories
		}
		m.mu.Unlock()
	}

	// Step 3: 重新获取 handlerInfo 和处理函数
	m.mu.RLock()
	hInfo = m.registries[handler]
	categories := hInfo.staticInfo[t]
	workerDoubleFunc := hInfo.workerDoubleFunc
	staticDoubleFunc := hInfo.staticDoubleFunc
	m.mu.RUnlock()

	if len(categories) == 0 || staticDoubleFunc == nil {
		return
	}

	// Step 4: 若有 workerDoubleFunc
	if workerDoubleFunc != nil {
		categories = workerDoubleFunc(t, categories, addValue, otherValue)
	}

	// Step 5: 调用 staticDoubleFunc
	staticDoubleFunc(t, categories, addValue, otherValue)
}





// ApplyStaticFunc 调用
func ApplyStaticFunc(handler StatisticHandler, t StatisticType, addValue int32) {
	GetGlobalManager().ApplyStaticFunc(handler, t, addValue)
}


// ApplyStaticFunc 调用
func ApplyStaticDoubleFunc(handler StatisticHandler, t StatisticType, addValue int32, otherValue int32) {
	GetGlobalManager().ApplyStaticDoubleFunc(handler, t, addValue, otherValue)
}