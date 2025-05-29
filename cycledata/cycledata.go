package cycledata

import (
	"log"
	"sync"
	"time"
)

/*
 * 类型定义
 */
type (
	UserID    int64
	TypeKey   int32
	CycleType string
)

/*
 * 周期类型 CycleType
 * 表示不同的时间周期分类，用于管理玩家数据的生命周期
 */
const (
	// DailyCycle 每日周期，每日自动重置
	DailyCycle CycleType = "daily"

	// WeeklyCycle 每周周期，每周自动重置
	WeeklyCycle CycleType = "weekly"

	// MonthlyCycle 每月周期，每月自动重置
	MonthlyCycle CycleType = "monthly"

	// LiftTime 永久周期，数据不重置
	LiftTime CycleType = "lifetime"

	// Newbie 新手周期，通常用于注册后一段时间内的统计
	Newbie CycleType = "newbie"

	// LimitTime 限时周期，适用于活动类限时周期数据
	LimitTime CycleType = "limitTime"

	// LoopTime 循环周期，可自定义循环周期逻辑（如每3天、每10局）
	LoopTime CycleType = "loopTime"
)

/*
 * 注册器变量
 */
var (
	/* 数据加载器映射：周期 -> 类型 -> 加载函数 */
	loaders = make(map[CycleType]map[TypeKey]func(cycle CycleType, typeKey TypeKey, userID UserID) *PlayerData)

	/* 数据创建器映射：周期 -> 类型 -> 创建函数 */
	creators = make(map[CycleType]map[TypeKey]func(userID UserID) *PlayerData)

	/* 数据存储函数 */
	stores = make(map[CycleType]map[TypeKey]func(cycle CycleType, typeKey TypeKey, data *PlayerData) error)

	/* 自定义过期处理函数 */
	cleanExpireds = make(map[CycleType]map[TypeKey]func(cycle CycleType, typeKey TypeKey, data *PlayerData))
)

func init() {
	loaders = make(map[CycleType]map[TypeKey]func(CycleType, TypeKey, UserID) *PlayerData)
	creators = make(map[CycleType]map[TypeKey]func(UserID) *PlayerData)
	stores = make(map[CycleType]map[TypeKey]func(CycleType, TypeKey, *PlayerData) error)
	cleanExpireds = make(map[CycleType]map[TypeKey]func(cycle CycleType, typeKey TypeKey, data *PlayerData))
}

/*
 * 玩家基础数据单元
 * 包含玩家基础数据和扩展字段
 */
type PlayerData struct {
	UserID     UserID
	UpdateTime time.Time
	ExpireTime int32
	MiscData   map[string]interface{}
	mu         sync.RWMutex
}

/*
 * 更新数据字段
 */
func (pd *PlayerData) update(key string, value interface{}) {
	pd.mu.Lock()
	defer pd.mu.Unlock()
	pd.MiscData[key] = value
	pd.UpdateTime = time.Now()
}

/*
 * 数据集合
 * 用于管理单个周期和类型下的所有玩家数据
 */
type dataCollection struct {
	mu   sync.RWMutex
	data map[UserID]*PlayerData
}

/*
 * 创建新的数据集合
 */
func newCollection() *dataCollection {
	return &dataCollection{
		data: make(map[UserID]*PlayerData),
	}
}

/*
 * 获取玩家数据（不存在则尝试通过注册的加载器/创建器构建）
 */
func (dc *dataCollection) get(cycle CycleType, typeKey TypeKey, userID UserID) *PlayerData {
	dc.mu.RLock()
	if data, ok := dc.data[userID]; ok {
		dc.mu.RUnlock()
		return data
	}
	dc.mu.RUnlock()

	dc.mu.Lock()
	defer dc.mu.Unlock()

	// 二次检查
	if data, ok := dc.data[userID]; ok {
		return data
	}

	// 加载器
	if loader := getLoader(cycle, typeKey); loader != nil {
		if loaded := loader(cycle, typeKey, userID); loaded != nil {
			dc.data[userID] = loaded
			return loaded
		}
	}

	// 创建器
	if creator := getCreator(cycle, typeKey); creator != nil {
		created := creator(userID)
		if created != nil {
			dc.data[userID] = created
			return created
		}
	}

	return nil
}

/*
 * 设置玩家数据（使用注册创建器，并注入 MiscData）
 */
func (dc *dataCollection) set(cycle CycleType, typeKey TypeKey, userID UserID, miscData map[string]interface{}) bool {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	// 如果已存在，则直接更新 MiscData
	if existing, ok := dc.data[userID]; ok {
		existing.MiscData = miscData
		return true
	}

	// 使用注册的创建器构造新的 PlayerData
	if creator := getCreator(cycle, typeKey); creator != nil {
		created := creator(userID)
		if created != nil {
			created.MiscData = miscData
			dc.data[userID] = created
			return true
		}
	}

	return false
}

/*
 * 清理冷数据回收内存
 */
func (dc *dataCollection) cleanCoolData(now int32, cycle CycleType, typeKey TypeKey) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	// First check if there's a custom expiration handler for this cycle and type
	handler := getStore(cycle, typeKey)

	// If no handler exists and no data to process, return early
	if handler == nil || len(dc.data) == 0 {
		return
	}
	// 把 now(int32 秒) 转成 time.Time
	nowTime := time.Unix(int64(now), 0)
	// 定义阈值时间，例如 6 小时之前
	threshold := nowTime.Add(-1 * time.Hour)
	hotData := make(map[UserID]*PlayerData)
	for uid, data := range dc.data {
		data.mu.RLock()
		lastUpdate := data.UpdateTime
		data.mu.RUnlock()

		if lastUpdate.Before(threshold) {
			if err := handler(cycle, typeKey, data); err != nil {
				log.Printf("Failed to store cold data for user %d: %v", uid, err)
			}
		} else {
			hotData[uid] = data
		}
	}
	dc.data = hotData
}

/*
 * 清理过期数据
 */
func (dc *dataCollection) cleanExpired(now int32, cycle CycleType, typeKey TypeKey) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	// First check if there's a custom expiration handler for this cycle and type
	handler := getCleanExpired(cycle, typeKey)

	// If no handler exists and no data to process, return early
	if handler == nil || len(dc.data) == 0 {
		return
	}
	isDelete := true
	for uid, data := range dc.data {
		// Default expiration logic
		if data.ExpireTime == 0 {
			if isDelete {
				isDelete = false
			}
			continue
		}
		if handler != nil {
			// Use custom expiration handler
			handler(cycle, typeKey, data)
			delete(dc.data, uid)
			continue
		}
		if data.ExpireTime <= now {
			delete(dc.data, uid)
		}
	}
	if isDelete {
		dc.data = make(map[UserID]*PlayerData)
	}
}

/*
 * 将集合中所有数据刷入存储器
 */
func (dc *dataCollection) flushAll(cycle CycleType, typeKey TypeKey) {
	if stores == nil {
		return
	}

	dc.mu.RLock()
	defer dc.mu.RUnlock()

	for _, data := range dc.data {
		// 创建器
		if store := getStore(cycle, typeKey); store != nil {
			if err := store(cycle, typeKey, data); err != nil {
				log.Printf("Failed to store data for cycle %v, type %v: %v", cycle, typeKey, err)
			}
		}
	}
}

/*
 * 周期服务
 * 用于管理某一周期下多个类型的数据集合
 */
type cycleService struct {
	mu            sync.RWMutex
	collections   map[TypeKey]*dataCollection
	defaultExpire int32
}

/*
 * 创建周期服务实例
 */
func newService(expire int32) *cycleService {
	return &cycleService{
		collections:   make(map[TypeKey]*dataCollection),
		defaultExpire: expire,
	}
}

/*
 * 获取指定类型的数据集合
 */
func (cs *cycleService) getCollection(typeKey TypeKey) *dataCollection {
	cs.mu.RLock()
	col, exists := cs.collections[typeKey]
	cs.mu.RUnlock()

	if exists {
		return col
	}

	cs.mu.Lock()
	defer cs.mu.Unlock()

	if col, exists = cs.collections[typeKey]; exists {
		return col
	}

	col = newCollection()
	cs.collections[typeKey] = col
	return col
}

/*
 * 刷新指定 typeKey 的数据
 */
func (cs *cycleService) flush(typeKey TypeKey, cycle CycleType) {
	cs.mu.RLock()
	col, ok := cs.collections[typeKey]
	cs.mu.RUnlock()
	if ok {
		col.flushAll(cycle, typeKey)
	}
}

/*
 * 刷新所有类型的数据
 */
func (cs *cycleService) flushAll(cycle CycleType) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	for typeKey, col := range cs.collections {
		col.flushAll(cycle, typeKey)
	}
}

/*
 * 全局周期处理器
 * 管理不同周期的服务及其定期任务
 */
type cycleHandler struct {
	mu       sync.RWMutex
	services map[CycleType]*cycleService
}

/*
 * 创建周期处理器
 */
func newCycleHandler() *cycleHandler {
	h := &cycleHandler{
		services: make(map[CycleType]*cycleService),
	}
	h.initPeriodicTasks()
	return h
}

/*
 * 启动定期冷数据清理任务
 */
func (h *cycleHandler) initPeriodicTasks() {
	go h.startCleanupRoutine()
}

/*
 * 每小时清理一次过期数据
 * 通过定时器触发，遍历所有周期类型，执行对应的过期数据清理操作
 * 过程先复制周期列表，避免长时间持锁，提升并发性能和安全性
 */
func (h *cycleHandler) startCleanupRoutine() {
	ticker := time.NewTicker(1 * time.Hour)
	for range ticker.C {
		// 读取锁保护访问 services map
		h.mu.RLock()
		// 复制所有周期类型的切片，避免持锁时间过长
		cycles := make([]CycleType, 0, len(h.services))
		for cycle := range h.services {
			cycles = append(cycles, cycle)
		}
		h.mu.RUnlock()

		// 解锁后逐个清理对应周期的过期数据，避免持锁时间过长阻塞其他操作
		for _, cycle := range cycles {
			h.cleanExpiredData(cycle)
		}
	}
}

/*
 * cleanExpiredDataByType 根据传入的周期 CycleType 和类型 TypeKey，清理对应数据集合中的过期数据
 *
 * 1. 获取对应周期的 service
 * 2. 获取对应类型的数据集合
 * 3. 获取当前时间戳
 * 4. 调用数据集合的 cleanExpired 方法
 */
func (h *cycleHandler) cleanExpiredDataByType(cycle CycleType, typeKey TypeKey) {
	// 获取对应周期的 service
	h.mu.RLock()
	service, ok := h.services[cycle]
	h.mu.RUnlock()

	if !ok || service == nil {
		return
	}

	// 获取对应类型的数据集合
	service.mu.RLock()
	col, ok := service.collections[typeKey]
	service.mu.RUnlock()

	if !ok || col == nil {
		return
	}

	// 获取当前时间戳并清理过期数据
	now := time.Now()
	timestamp := int32(now.Unix())
	col.cleanExpired(timestamp, cycle, typeKey)
}

/*
 * CleanExpiredDataByType 公开方法，清理指定周期和类型的过期数据
 */
func CleanExpiredDataByType(cycle CycleType, typeKey TypeKey) {
	globalHandler.cleanExpiredDataByType(cycle, typeKey)
}

/*
 * cleanExpiredData 根据传入的周期 CycleType，清理对应周期服务中的过期数据
 *
 * 1. 先对 cycleHandler 的服务 map 加读锁，获取对应周期的 service 指针
 * 2. 解锁，避免长时间持锁影响并发
 * 3. 如果对应周期的 service 不存在，直接返回
 * 4. 获取当前时间，作为过期判断依据
 * 5. 对 service 内部的 collections 加读锁，遍历所有 TypeKey 对应的数据集合
 * 6. 调用各集合的 cleanExpired 方法，执行具体的过期清理逻辑
 */
func (h *cycleHandler) cleanExpiredData(cycle CycleType) {
	// 加读锁读取指定周期的 service
	h.mu.RLock()
	service, ok := h.services[cycle]
	h.mu.RUnlock()

	if !ok || service == nil {
		// 没有对应周期的服务则不做任何处理
		return
	}

	now := time.Now()
	timestamp := int32(now.Unix())

	// 加读锁访问 service 内部 collections
	service.mu.RLock()
	defer service.mu.RUnlock()

	// 遍历所有 TypeKey 对应的数据集合，执行过期清理
	for typeKey, col := range service.collections {
		col.cleanExpired(timestamp, cycle, typeKey)
	}
}

/*
* cleanCoolData 根据传入的周期 CycleType，清理对应周期服务中的过期数据
*
* 1. 先对 cycleHandler 的服务 map 加读锁，获取对应周期的 service 指针
* 2. 解锁，避免长时间持锁影响并发
* 3. 如果对应周期的 service 不存在，直接返回
* 4. 获取当前时间，作为过期判断依据
* 5. 对 service 内部的 collections 加读锁，遍历所有 TypeKey 对应的数据集合
* 6. 调用各集合的 cleanExpired 方法，执行具体的过期清理逻辑
 */
func (h *cycleHandler) cleanCoolData(cycle CycleType) {
	// 加读锁读取指定周期的 service
	h.mu.RLock()
	service, ok := h.services[cycle]
	h.mu.RUnlock()

	if !ok || service == nil {
		// 没有对应周期的服务则不做任何处理
		return
	}

	now := time.Now()
	timestamp := int32(now.Unix())

	// 加读锁访问 service 内部 collections
	service.mu.RLock()
	defer service.mu.RUnlock()

	// 遍历所有 TypeKey 对应的数据集合，执行过期清理
	for typeKey, col := range service.collections {
		col.cleanCoolData(timestamp, cycle, typeKey)
	}
}

/*
 * 获取指定周期服务（自动初始化）
 */
func (h *cycleHandler) getService(cycle CycleType, expire int32) *cycleService {
	h.mu.RLock()
	s, exists := h.services[cycle]
	h.mu.RUnlock()

	if exists {
		return s
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	if s, exists = h.services[cycle]; exists {
		return s
	}

	s = newService(expire)
	h.services[cycle] = s
	return s
}

/*
 * 刷新指定周期和类型的所有玩家数据
 */
func Flush(cycle CycleType, typeKey TypeKey) {
	globalHandler.getService(cycle, DefaultExpireFor(cycle, typeKey)).
		flush(typeKey, cycle)
}

/*
 * 刷新所有周期、类型、用户数据
 */
func FlushAll() {
	globalHandler.mu.RLock()
	defer globalHandler.mu.RUnlock()

	for cycle, service := range globalHandler.services {
		service.flushAll(cycle)
	}
}

/*
 * 全局周期处理器实例
 */
var globalHandler = newCycleHandler()

/*
 * 获取指定加载器
 */
func getLoader(cycle CycleType, typeKey TypeKey) func(CycleType, TypeKey, UserID) *PlayerData {
	if m, ok := loaders[cycle]; ok {
		return m[typeKey]
	}
	return nil
}

/*
 * 获取指定创建器
 */
func getCreator(cycle CycleType, typeKey TypeKey) func(UserID) *PlayerData {
	if m, ok := creators[cycle]; ok {
		return m[typeKey]
	}
	return nil
}

/*
 * 获取指定存储器
 */
// getStore retrieves the appropriate store function for the given cycle and type
func getStore(cycle CycleType, typeKey TypeKey) func(CycleType, TypeKey, *PlayerData) error {
	if typeStores, ok := stores[cycle]; ok {
		if store, ok := typeStores[typeKey]; ok {
			return store
		}
	}
	return nil
}

/*
 * 获取指定过期处理函数
 */
func getCleanExpired(cycle CycleType, typeKey TypeKey) func(cycle CycleType, typeKey TypeKey, data *PlayerData) {
	if m, ok := cleanExpireds[cycle]; ok {
		return m[typeKey]
	}
	return nil
}
