package cache

import (
	"sync"
	"time"
)

type FlushHandler[T any] interface {
	Flush(data []T) error
}

type CacheService[T any] struct {
	handler  FlushHandler[T]
	interval time.Duration

	mutex   sync.Mutex
	cache   []T
	stopCh  chan struct{}
	started bool
}

func NewCacheService[T any](handler FlushHandler[T], interval time.Duration) *CacheService[T] {
	return &CacheService[T]{
		handler:  handler,
		interval: interval,
		cache:    make([]T, 0),
		stopCh:   make(chan struct{}),
	}
}

func (s *CacheService[T]) Start() {
	if s.started {
		return
	}
	s.started = true
	go s.run()
}

func (s *CacheService[T]) Stop() {
	if !s.started {
		return
	}
	close(s.stopCh)
	s.started = false
}

func (s *CacheService[T]) Push(data T) {
	s.mutex.Lock()
	s.cache = append(s.cache, data)
	s.mutex.Unlock()
}

func (s *CacheService[T]) run() {
    ticker := time.NewTicker(s.interval)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            s.flush()
        case _, ok := <-s.stopCh:  // 检查通道是否已关闭
            if !ok {
                s.flush()
                return
            }
        }
    }
}
func (s *CacheService[T]) flush() {
	s.mutex.Lock()
	data := s.cache
	s.cache = make([]T, 0)
	s.mutex.Unlock()
	if len(data) > 0 && s.handler != nil {
		_ = s.handler.Flush(data)
	}
}
