package cache

import (
	"sync"
	"time"
)

type ListFlushHandler interface {
	Flush(data []interface{}) error
}

type CacheService struct {
	handler  ListFlushHandler
	interval time.Duration

	mutex   sync.Mutex
	cache   []interface{}
	stopCh  chan struct{}
	started bool
}

func NewCacheService(handler ListFlushHandler, interval time.Duration) *CacheService {
	return &CacheService{
		handler:  handler,
		interval: interval,
		cache:    make([]interface{}, 0),
		stopCh:   make(chan struct{}),
	}
}

func (s *CacheService) Start() {
	if s.started {
		return
	}
	s.started = true
	go s.run()
}

func (s *CacheService) Stop() {
	if !s.started {
		return
	}
	close(s.stopCh)
	s.started = false
}

func (s *CacheService) Push(data interface{}) {
	s.mutex.Lock()
	s.cache = append(s.cache, data)
	s.mutex.Unlock()
}

// todo 调整为 注册对应定时器
func (s *CacheService) run() {
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

func (s *CacheService) flush() {
	s.mutex.Lock()
	data := s.cache
	s.cache = make([]interface{}, 0)
	s.mutex.Unlock()
	if len(data) > 0 && s.handler != nil {
		_ = s.handler.Flush(data)
	}
}