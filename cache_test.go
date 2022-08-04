package cache

import (
    "sync"
    "testing"
    "time"
)

func mockTime(sec int64) {
    GetTime = func() time.Time {
        return time.Unix(sec, 0)
    }
}
func TestNewCache(t *testing.T) {
    mockTime(0)
    c := NewCache(time.Second, 0.9, false)
    c.OnRemove(func(key string, value interface{}) {
        t.Log("delete", key, value)
    })

    c.Put("k1", "v1", time.Second*10)

    mockTime(8)
    t.Log(c.Get("k1"))

    mockTime(20)

    c.Put("k1", "vv1", time.Second)
    t.Log(c.Get("k1"))

    mockTime(21)
    time.Sleep(time.Second)
    t.Log(c.Get("k1"))

    c.Clear()
}
func TestBenchCache(t *testing.T) {
    c := NewCache(time.Minute, 1, true)
    var (
        wg    sync.WaitGroup
        count int32
    )
    for i := 0; i < 1000; i++ {
        wg.Add(1)
        go testWriteCache(c, &count, &wg)
    }
    wg.Wait()

    t.Log("qps:", count/5)
}
