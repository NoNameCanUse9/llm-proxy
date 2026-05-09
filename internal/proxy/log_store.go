package proxy

import (
	"sync"
)

var (
	GlobalLogStore = NewLogStore(500) // 默认内存保留最近 500 条
)

type LogStore struct {
	mu       sync.RWMutex
	logs     []interface{} // 使用 interface{} 兼容不同的 Log 类型
	maxSize  int
}

func NewLogStore(maxSize int) *LogStore {
	return &LogStore{
		logs:    make([]interface{}, 0),
		maxSize: maxSize,
	}
}

func (s *LogStore) Add(log interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.logs = append([]interface{}{log}, s.logs...) // 最新的放在前面
	if len(s.logs) > s.maxSize {
		s.logs = s.logs[:s.maxSize]
	}
}

func (s *LogStore) GetAll() []interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	// 返回副本
	res := make([]interface{}, len(s.logs))
	copy(res, s.logs)
	return res
}
