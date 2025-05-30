package cache

import (
	"log"
	"sync"
	"time"
)

// CacheServiceRegistry 缓存服务注册中心
type CacheServiceRegistry struct {
	mu           sync.RWMutex
	kvServices   map[string]*KVCacheService
	listServices map[string]*CacheService
	kvHandlers   map[string]KVFlushHandler
	listHandlers map[string]ListFlushHandler
}

var registry = &CacheServiceRegistry{
	kvServices:   make(map[string]*KVCacheService),
	listServices: make(map[string]*CacheService),
	kvHandlers:   make(map[string]KVFlushHandler),
	listHandlers: make(map[string]ListFlushHandler),
}

// RegisterKVService 注册KV缓存服务（初始化延迟到 Start）
func RegisterKVService(id string, handler KVFlushHandler, interval time.Duration, initializer func()) {
	registry.mu.Lock()
	defer registry.mu.Unlock()

	// 创建 KV 缓存服务
	kvService := NewKVCacheService(handler, interval)
	registry.kvServices[id] = kvService
	registry.kvHandlers[id] = handler

	// 注册模块（暂不初始化）
	module := &KeyCacheModule{
		ID:          id,
		Cache:       kvService,
		initializer: initializer, // 保存初始化函数
	}
	GetServer().RegisterModule(module)
}

// RegisterListService 注册列表缓存服务（初始化延迟到 Start）
func RegisterListService(id string, handler ListFlushHandler, interval time.Duration, initializer func()) {
	registry.mu.Lock()
	defer registry.mu.Unlock()

	// 创建 List 缓存服务
	listService := NewCacheService(handler, interval)
	registry.listServices[id] = listService
	registry.listHandlers[id] = handler

	// 注册模块（暂不初始化）
	module := &ListCacheModule{
		ID:          id,
		Cache:       listService,
		initializer: initializer, // 保存初始化函数
	}
	GetServer().RegisterModule(module)
}

// KeyCacheModule KV缓存模块
type KeyCacheModule struct {
	ID          string
	Cache       *KVCacheService
	initializer func() // 新增：保存初始化函数
}

// Init 初始化模块（私有方法，仅供 Start 调用）
func (m *KeyCacheModule) init() {
	if m.initializer != nil {
		m.initializer()
		log.Printf("Initialized KV cache module: %s", m.ID)
	}
}

func (m *KeyCacheModule) Start() error {
	registry.mu.RLock()
	defer registry.mu.RUnlock()

	// 1. 先初始化
	m.init()

	// 2. 再启动服务
	if service, ok := registry.kvServices[m.ID]; ok {
		m.Cache = service
		m.Cache.Start()
		log.Printf("KeyCacheModule %s started", m.ID)
	}
	return nil
}

func (m *KeyCacheModule) Stop() error {
	log.Printf("KeyCacheModule %s stopped", m.ID)
	if m.Cache != nil {
		m.Cache.Stop()
	}
	return nil
}

func (m *KeyCacheModule) Name() string {
	return m.ID
}

func (m *KeyCacheModule) Push(data interface{}) error {
	if m.Cache != nil {
		return m.Cache.Push(data) // Call the service's method
	}
	return nil
}

func (m *KeyCacheModule) UpdateKeyValue(key int64, data interface{}) error {
	if m.Cache != nil {
		return m.Cache.UpdateKeyValue(key, data) // Call the service's method
	}
	return nil
}

func (m *KeyCacheModule) GetKeyValue(key int64) interface{} {
	if m.Cache != nil {
		return m.Cache.GetKeyValue(key) // Call the service's method
	}
	return nil
}

// ListCacheModule 列表缓存模块
type ListCacheModule struct {
	ID          string
	Cache       *CacheService
	initializer func() // 新增：保存初始化函数
}

// Init 初始化模块（私有方法，仅供 Start 调用）
func (m *ListCacheModule) init() {
	if m.initializer != nil {
		m.initializer()
		log.Printf("Initialized List cache module: %s", m.ID)
	}
}

func (m *ListCacheModule) Start() error {
	registry.mu.RLock()
	defer registry.mu.RUnlock()

	// 1. 先初始化
	m.init()

	// 2. 再启动服务
	if service, ok := registry.listServices[m.ID]; ok {
		m.Cache = service
		m.Cache.Start()
		log.Printf("ListCacheModule %s started", m.ID)
	}
	return nil
}

func (m *ListCacheModule) Stop() error {
	log.Printf("ListCacheModule %s stopped", m.ID)
	if m.Cache != nil {
		m.Cache.Stop()
	}
	return nil
}

func (m *ListCacheModule) Name() string {
	return m.ID
}

func (m *ListCacheModule) Push(data interface{}) error {
	if m.Cache != nil {
		return m.Cache.Push(data) // Call the service's method
	}
	return nil
}

func (m *ListCacheModule) UpdateKeyValue(key int64, data interface{}) error {
	if m.Cache != nil {
		return m.Cache.UpdateKeyValue(key, data) // Call the service's method
	}
	return nil
}

func (m *ListCacheModule) GetKeyValue(key int64) interface{} {
	if m.Cache != nil {
		return m.Cache.GetKeyValue(key) // Call the service's method
	}
	return nil
}
