package config

import (
	"encoding/json"
	"sync"
)

// AdmisRequestResult 定义返回的处理结果
type AdmisRequestResult struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

type AdmisRequestResultObj struct {
	Type      string           `json:"type"`
	RequestId string           `json:"request_id"`
	DeployRes *json.RawMessage `json:"deploy_res"`
}

var (
	RequestFutures = NewRequestFuturesMap()
)

// RequestFuturesMap 是一个封装了 map 和同步机制的类型
type RequestFuturesMap struct {
	mu      sync.RWMutex                     // 读写锁
	futures map[string]chan *json.RawMessage // map 存储的实际数据
}

// NewRequestFuturesMap 创建一个新的 RequestFuturesMap 实例
func NewRequestFuturesMap() *RequestFuturesMap {
	return &RequestFuturesMap{
		futures: make(map[string]chan *json.RawMessage, 10),
	}
}

// Set 向 map 中添加一个键值对，使用写锁确保线程安全
func (r *RequestFuturesMap) Set(key string, value chan *json.RawMessage) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.futures[key] = value
}

// Get 从 map 中获取一个值，使用读锁确保线程安全
func (r *RequestFuturesMap) Get(key string) (chan *json.RawMessage, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	value, exists := r.futures[key]
	return value, exists
}

// Delete 从 map 中删除一个键值对，使用写锁确保线程安全
func (r *RequestFuturesMap) Delete(key string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.futures, key)
}

// Size 获取 map 的大小，使用读锁确保线程安全
func (r *RequestFuturesMap) Size() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.futures)
}
