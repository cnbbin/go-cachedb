package cache

import (
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"context"
)

// Module 定义模块统一接口
type Module interface {
	Start() error
	Stop() error
	Name() string
}

// Server 结构体，类似 grpc.Server
type Server struct {
	Modules map[string]Module
	mutex   sync.RWMutex
}

var (
	srv *Server
	once sync.Once
)

// NewCacheServer 构造函数
func NewCacheServer() *Server {
	once.Do(func() {
		srv = &Server{
			Modules: make(map[string]Module),
		}
	})
	return srv
}

// GetServer 获取服务器实例 (单例模式)
func GetServer() *Server {
	once.Do(func() {
		srv = NewCacheServer()
	})
	return srv
}

// RegisterModule 注册一个服务模块
func (s *Server) RegisterModule(m Module) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.Modules[m.Name()] = m
}

// GetModule 获取已注册的模块
func (s *Server) GetModule(name string) (Module, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	module, exists := s.Modules[name]
	return module, exists
}

// Start 启动所有模块
func (s *Server) Start(ctx context.Context) error {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	log.Printf("[Server] Started all module: ")

	for name, mod := range s.Modules {
		log.Printf("[Server] Started module: %s", name)
		if err := mod.Start(); err != nil {
			return err
		}
	}
	return nil
}

// Stop 停止所有模块
func (s *Server) Stop(ctx context.Context) error {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for name, mod := range s.Modules {
		log.Printf("[Server] Stopping module: %s", name)
		if err := mod.Stop(); err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) Register(name string, m Module) {
	if name != m.Name() {
		log.Printf("[Server] Warning: Registered module name mismatch: key=%s actual=%s", name, m.Name())
	}
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.Modules[name] = m
	log.Printf("[Server] Registered module: %s", name)
}

// RegisterModules 批量注册模块
func (s *Server) RegisterModules(mods ...Module) {
	for _, m := range mods {
		s.Register(m.Name(), m)
	}
}

// Serve 启动并监听关闭信号
func (s *Server) Serve() {
	ctx := context.Background()

	if err := s.Start(ctx); err != nil {
		log.Fatalf("[Server] Start failed: %v", err)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop
	log.Println("[Server] Shutting down...")
	s.Stop(ctx)
	log.Println("[Server] Exit.")
}

// init 初始化函数
func init() {
	// 初始化服务器实例
	srv = NewCacheServer()
	log.Println("[Server] Initialized cache server")
}