package cache

import (
	"sync"
	"time"
)

type KVFlushHandler interface {
	Flush(data []interface{}) error
}

type KVCacheService struct {
	handler  KVFlushHandler
	interval time.Duration

	mutex   sync.Mutex
	cache   map[interface{}]interface{}
	stopCh  chan struct{}
	started bool
}

func NewKVCacheService(handler KVFlushHandler, interval time.Duration) *KVCacheService {
	return &KVCacheService{
		handler:  handler,
		interval: interval,
		cache:    make(map[interface{}]interface{}),
		stopCh:   make(chan struct{}),
	}
}

func (s *KVCacheService) Start() {
	if s.started {
		return
	}
	s.started = true
	go s.run()
}

func (s *KVCacheService) Stop() {
	if !s.started {
		return
	}
	close(s.stopCh)
	s.started = false
}

func (s *KVCacheService) UpdateKeyValue(key interface{}, value interface{}) {
	s.mutex.Lock()
	s.cache[key] = value
	s.mutex.Unlock()
}

func (s *KVCacheService) run() {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.flush()
		case _, ok := <-s.stopCh:
			if !ok {
				s.flush()
				return
			}
		}
	}
}

func (s *KVCacheService) flush() {
	s.mutex.Lock()
	data := make([]interface{}, 0, len(s.cache))
	for _, v := range s.cache {
		data = append(data, v)
	}
	s.cache = make(map[interface{}]interface{})
	s.mutex.Unlock()
	if len(data) > 0 && s.handler != nil {
		_ = s.handler.Flush(data)
	}
}