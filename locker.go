package cache

import "sync"

type (
    mKeyLocker struct {
        sync.Mutex
        count int
    }
    Locker struct {
        lock  sync.Mutex
        locks map[string]*mKeyLocker
    }
)

func NewLocker() *Locker {
    return &Locker{
        lock:  sync.Mutex{},
        locks: make(map[string]*mKeyLocker),
    }
}

func (l *Locker) Lock(key string) {
    l.lock.Lock()

    keyLocker := l.locks[key]
    if keyLocker == nil {
        keyLocker = &mKeyLocker{}
        l.locks[key] = keyLocker
    }
    keyLocker.count++
    l.lock.Unlock()

    keyLocker.Lock()
}
func (l *Locker) UnLock(key string) {
    l.lock.Lock()

    keyLocker := l.locks[key]
    keyLocker.Unlock()
    keyLocker.count--
    if keyLocker.count == 0 {
        delete(l.locks, key)
    }

    l.lock.Unlock()
}

var (
    _global = NewLocker()
    _       = Lock
    _       = Unlock
)

func Lock(key string) {
    _global.Lock(key)
}
func Unlock(key string) {
    _global.UnLock(key)
}
