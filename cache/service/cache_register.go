package cache

import (
	"log"
	"sync"
	"time"
)

// CacheServiceRegistry 缓存服务注册中心
type CacheServiceRegistry struct {
	mu               sync.RWMutex
	kvServices       map[string]*KVCacheService[any]
	listServices     map[string]*CacheService[any]
	kvInitializers   map[string]func()
	listInitializers map[string]func()
	kvHandlers       map[string]FlushHandler[any]
	listHandlers     map[string]FlushHandler[any]
}

var registry = &CacheServiceRegistry{
	kvServices:       make(map[string]*KVCacheService[any]),
	listServices:     make(map[string]*CacheService[any]),
	kvInitializers:   make(map[string]func()),
	listInitializers: make(map[string]func()),
	kvHandlers:       make(map[string]FlushHandler[any]),
	listHandlers:     make(map[string]FlushHandler[any]),
}

// RegisterKVService 注册KV缓存服务
func RegisterKVService(id string, handler FlushHandler[any], interval time.Duration, initializer func()) {
	registry.mu.Lock()
	defer registry.mu.Unlock()

	registry.kvServices[id] = NewKVCacheService[any](handler, interval)
	registry.kvInitializers[id] = initializer
	registry.kvHandlers[id] = handler
}

// RegisterListService 注册列表缓存服务
func RegisterListService(id string, handler FlushHandler[any], interval time.Duration, initializer func()) {
	registry.mu.Lock()
	defer registry.mu.Unlock()

	registry.listServices[id] = NewCacheService[any](handler, interval)
	registry.listInitializers[id] = initializer
	registry.listHandlers[id] = handler
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
	Cache *KVCacheService[any]
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
	Cache *CacheService[any]
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

// 初始化函数 - 在程序启动时调用
func init() {
	// 注册所有KV缓存服务
	RegisterKVService("kvPlayerClothes", &PlayerClothesFlushHandler{}, 1*time.Second, InitPlayerClothes)
	// 注册列表缓存服务
	RegisterListService("listPlayerCurrency", &MyHandler{}, 5*time.Second, InitPlayerClothes)
}