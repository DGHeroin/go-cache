package cache

import (
    "fmt"
    "sync"
    "time"
)

type (
    ShardCache struct {
        sync.RWMutex
        hash  *Hash
        peers map[string]*TimeBaseCache
    }
)

func (s *ShardCache) Foreach(f func(key string, value interface{}) bool) {
    for _, cache := range s.peers {
        cache.Foreach(f)
    }
}
func NewShardCache(n int, GCDuration time.Duration, fluctuation01 float32, autoRenewal bool) *ShardCache {
    s := &ShardCache{
        hash:  NewHash(),
        peers: map[string]*TimeBaseCache{},
    }
    for i := 0; i < n; i++ {
        nodeName := fmt.Sprintf("node#%d", i+1)
        s.hash.Add(nodeName, 1)
        s.peers[nodeName] = NewCache(GCDuration, fluctuation01, autoRenewal)
    }
    return s
}
func (s *ShardCache) getNode(key string) *TimeBaseCache {
    s.RLock()
    defer s.RUnlock()
    nodeName := s.hash.GetNode(key)
    node := s.peers[nodeName]
    return node
}
func (s *ShardCache) Put(key string, value interface{}, duration time.Duration) {
    node := s.getNode(key)
    node.Put(key, value, duration)
}
func (s *ShardCache) Get(key string) (interface{}, bool) {
    node := s.getNode(key)
    return node.Get(key)
}
func (s *ShardCache) OnRemove(fn func(key string, value interface{})) {
    for _, cache := range s.peers {
        cache.OnRemove(fn)
    }
}
func (s *ShardCache) Clear() {
    for _, cache := range s.peers {
        cache.Clear()
    }
}
