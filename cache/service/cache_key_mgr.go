package cache


import (
	"sync"
	"time"
	"fmt"
)

type FlushKeyHandler[T any] interface {
	FlushKey(data []T) error
}

type KVCacheService[T any] struct {
	handler  FlushHandler[T]
	interval time.Duration

	mutex   sync.Mutex
	cache   map[any]T
	stopCh  chan struct{}
	started bool
}

func NewKVCacheService[T any](handler FlushHandler[T], interval time.Duration) *KVCacheService[T] {
	return &KVCacheService[T]{
		handler:  handler,
		interval: interval,
		cache:    make(map[any]T),
		stopCh:   make(chan struct{}),
	}
}

func (s *KVCacheService[T]) Start() {
	if s.started {
		return
	}
	s.started = true
	go s.run()
}

func (s *KVCacheService[T]) Stop() {
	if !s.started {
		return
	}
	close(s.stopCh)
	s.started = false
}

// 处理如果跟上次key一样
func (s *KVCacheService[T]) UpdateKeyValue(key any, value T) {
	s.mutex.Lock()
	s.cache[key] = value
	s.mutex.Unlock()
	fmt.Println("UpdateKeyValue", key, value)
}

func (s *KVCacheService[T]) run() {
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

func (s *KVCacheService[T]) flush() {
	s.mutex.Lock()
	data := make([]T, 0, len(s.cache))
	for _, v := range s.cache {
		data = append(data, v)
	}
	s.cache = make(map[any]T)
	s.mutex.Unlock()
	if len(data) > 0 && s.handler != nil {
		_ = s.handler.Flush(data)
	}
}