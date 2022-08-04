package cache

import "time"

type (
    Cache interface {
        Put(key string, value interface{}, duration time.Duration)
        Get(key string) (interface{}, bool)
        Clear()
        OnRemove(fn func(key string, value interface{}))
        Foreach(f func(key string, value interface{}) bool)
    }
)
