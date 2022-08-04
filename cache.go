package cache

import (
    "math/rand"
    "sync"
    "sync/atomic"
    "time"
)

type (
    TimeBaseCache struct {
        locker        *Locker
        container     sync.Map
        count         int64
        fluctuation01 float32
        isRunning     bool
        fluIndex      int
        autoRenewal   bool
        interval      time.Duration
        onRemove      func(key string, value interface{})
    }
    item struct {
        key      string
        data     interface{}
        expire   time.Time
        duration time.Duration
        access   time.Time
    }
)

func (c *TimeBaseCache) Count() int64 {
    return atomic.LoadInt64(&c.count)
}
func (c *TimeBaseCache) GC() {
    c.container.Range(func(k, v interface{}) bool {
        key := k.(string)
        value := v.(*item)
        c.locker.Lock(key)
        defer c.locker.UnLock(key)
        if value.IsTimeout() {
            c.onRemove(key, value.data)
            c.container.Delete(key)
            atomic.AddInt64(&c.count, -1)
        }
        return true
    })
}
func (c *TimeBaseCache) Get(key string) (interface{}, bool) {
    c.locker.Lock(key)
    defer c.locker.UnLock(key)

    p, ok := c.container.Load(key)
    if !ok {
        return nil, false
    }
    it := p.(*item)
    if it.IsTimeout() {
        return nil, false
    }
    return it.data, true
}
func (c *TimeBaseCache) addItemDuration(it *item, duration time.Duration, autoRenewal bool) {
    if !autoRenewal {
        return
    }
    if duration > 0 {
        if c.fluctuation01 > 0 { // 在给定失效时间左右偏移, 避免缓存雪崩
            duration -= time.Duration(rand.Int63n(int64(float64(duration) * float64(1-c.fluctuation01))))
        }
        it.expire = GetTime().Add(duration)
    }
}
func (c *TimeBaseCache) Put(key string, value interface{}, duration time.Duration) {
    c.locker.Lock(key)
    defer c.locker.UnLock(key)

    it := &item{
        key:      key,
        data:     value,
        duration: duration,
    }
    c.addItemDuration(it, duration, true)
    if p, ok := c.container.Load(key); ok {
        c.onRemove(key, p.(*item).data)
        atomic.AddInt64(&c.count, -1)
        c.container.Delete(key)
    }
    c.container.Store(key, it)
    atomic.AddInt64(&c.count, 1)

}
func (c *TimeBaseCache) Remove(k string) interface{} {
    c.locker.Lock(k)
    defer c.locker.UnLock(k)

    p, ok := c.container.LoadAndDelete(k)
    if !ok {
        return nil
    }
    atomic.AddInt64(&c.count, -1)
    return p.(*item).data
}

func (c *TimeBaseCache) Foreach(f func(key string, value interface{}) bool) {
    c.container.Range(func(k, v interface{}) bool {
        key := k.(string)
        value := v.(*item)
        c.locker.Lock(key)
        defer c.locker.UnLock(key)
        f(key, value.data)
        return true
    })
}

func (it *item) IsTimeout() bool {
    if it.duration == 0 {
        return false
    }
    return GetTime().After(it.expire)
}
func (c *TimeBaseCache) OnRemove(fn func(key string, value interface{})) {
    c.onRemove = fn
}

func (c *TimeBaseCache) Clear() {
    c.container.Range(func(k, v interface{}) bool {
        key := k.(string)
        value := v.(*item)
        c.locker.Lock(key)
        defer c.locker.UnLock(key)
        c.onRemove(key, value.data)
        c.container.Delete(key)
        atomic.AddInt64(&c.count, -1)
        return true
    })
}
func (c *TimeBaseCache) StopGC() {
    c.isRunning = false
}
func (c *TimeBaseCache) StartGC() {
    if c.interval == 0 {
        return
    }
    c.isRunning = true
    go func() {
        ticker := time.NewTicker(c.interval)
        defer ticker.Stop()
        for c.isRunning {
            select {
            case <-ticker.C:
                c.GC()
            }
        }
    }()
}
func NewCache(GCDuration time.Duration, fluctuation01 float32, autoRenewal bool) *TimeBaseCache {
    c := &TimeBaseCache{
        fluctuation01: MathClampF32(fluctuation01, 0, 1),
        autoRenewal:   autoRenewal,
        interval:      GCDuration,
        locker:        NewLocker(),
        onRemove:      func(key string, value interface{}) {},
    }
    if c.fluctuation01 > 0 {
        rand.Seed(GetTime().UnixNano())
    }
    c.StartGC()
    return c
}
