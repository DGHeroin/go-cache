package cache

import (
    "sync"
    "sync/atomic"
    "testing"
    "time"
)

func TestNewShardCache(t *testing.T) {
    c := NewShardCache(10, time.Second, 0.9, false)
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
func TestBenchShardCache(t *testing.T) {
    c := NewShardCache(10, time.Minute, 1, true)
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

func testWriteCache(c Cache, counter *int32, wg *sync.WaitGroup) {
    defer wg.Done()
    startTime := time.Now()

    for {
        c.Put("k1", "v1", 0)
        atomic.AddInt32(counter, 1)
        if time.Since(startTime) > time.Second*5 {
            break
        }
    }
}
