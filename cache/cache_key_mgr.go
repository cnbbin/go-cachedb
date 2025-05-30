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

func (s *KVCacheService) UpdateKeyValue(key int64, value interface{}) error {
	s.mutex.Lock()
	s.cache[key] = value
	s.mutex.Unlock()
	return nil
}

func (s *KVCacheService) GetKeyValue(key int64) (value interface{}) {
	s.mutex.Lock()
	value, valueExist := s.cache[key]
	s.mutex.Unlock()
	if !valueExist {
		return nil
	}
	return value
}

func (s *KVCacheService) Push(data interface{}) error {
	return nil
}

// todo 调整为 注册对应定时器
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
