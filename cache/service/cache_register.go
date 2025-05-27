package cache

import (
	"log"
	"sync"
	"time"
)

// CacheServiceRegistry 缓存服务注册中心
type CacheServiceRegistry struct {
	mu               sync.RWMutex
	kvServices       map[string]*KVCacheService
	listServices     map[string]*CacheService
	kvInitializers   map[string]func()
	listInitializers map[string]func()
	kvHandlers       map[string]KVFlushHandler
	listHandlers     map[string]ListFlushHandler
}

var registry = &CacheServiceRegistry{
	kvServices:       make(map[string]*KVCacheService),
	listServices:     make(map[string]*CacheService),
	kvInitializers:   make(map[string]func()),
	listInitializers: make(map[string]func()),
	kvHandlers:       make(map[string]KVFlushHandler),
	listHandlers:     make(map[string]ListFlushHandler),
}

// RegisterKVService 注册KV缓存服务
func RegisterKVService(id string, handler KVFlushHandler, interval time.Duration, initializer func()) {
	registry.mu.Lock()
	defer registry.mu.Unlock()
	// 创建 KV 缓存服务
	kvService := NewKVCacheService(handler, interval)
	registry.kvServices[id] = kvService
	registry.kvInitializers[id] = initializer
	registry.kvHandlers[id] = handler

	// 自动注册到 cache.Server
	module := &KeyCacheModule{
		ID:    id,
		Cache: kvService,
	}
	GetServer().RegisterModule(module)
}

// RegisterListService 注册列表缓存服务
func RegisterListService(id string, handler ListFlushHandler, interval time.Duration, initializer func()) {
	registry.mu.Lock()
	defer registry.mu.Unlock()

	// 创建 List 缓存服务
	listService := NewCacheService(handler, interval)
	registry.listServices[id] = listService
	registry.listInitializers[id] = initializer
	registry.listHandlers[id] = handler

	// 自动注册到 cache.Server
	module := &ListCacheModule{
		ID:    id,
		Cache: listService,
	}
	GetServer().RegisterModule(module)
}

// InitializeServices 初始化所有注册的服务
func InitializeServices() {
	registry.mu.RLock()
	defer registry.mu.RUnlock()

	// 初始化KV服务
	for id, initializer := range registry.kvInitializers {
		if initializer != nil {
			initializer()
		}
		log.Printf("Initialized KV cache service: %s", id)
	}

	// 初始化列表服务
	for id, initializer := range registry.listInitializers {
		if initializer != nil {
			initializer()
		}
		log.Printf("Initialized List cache service: %s", id)
	}
}

// KeyCacheModule KV缓存模块
type KeyCacheModule struct {
	ID    string
	Cache *KVCacheService
}

func (d *KeyCacheModule) Start() error {
	registry.mu.RLock()
	defer registry.mu.RUnlock()

	if service, ok := registry.kvServices[d.ID]; ok {
		d.Cache = service
		d.Cache.Start()
		log.Printf("KeyCacheModule %s started", d.ID)
		return nil
	}
	return nil
}

func (d *KeyCacheModule) Stop() error {
	log.Printf("KeyCacheModule %s stopped", d.ID)
	if d.Cache != nil {
		d.Cache.Stop()
	}
	return nil
}

func (d *KeyCacheModule) Name() string {
	return "kvcache:" + d.ID
}

// ListCacheModule 列表缓存模块
type ListCacheModule struct {
	ID    string
	Cache *CacheService
}

func (d *ListCacheModule) Start() error {
	registry.mu.RLock()
	defer registry.mu.RUnlock()

	if service, ok := registry.listServices[d.ID]; ok {
		d.Cache = service
		d.Cache.Start()
		log.Printf("ListCacheModule %s started", d.ID)
		return nil
	}
	return nil
}

func (d *ListCacheModule) Stop() error {
	log.Printf("ListCacheModule %s stopped", d.ID)
	if d.Cache != nil {
		d.Cache.Stop()
	}
	return nil
}

func (d *ListCacheModule) Name() string {
	return "listcache:" + d.ID
}