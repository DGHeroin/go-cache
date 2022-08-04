package cache

import (
    "hash/crc32"
    "sort"
    "strconv"
    "sync"
)

type (
    Hash struct {
        sortedNodes []uint32
        circle      map[uint32]string
        nodes       map[string]bool
        vCount      int
        sync.RWMutex
    }
)

func NewHash() *Hash {
    c := &Hash{
        circle: map[uint32]string{},
        nodes:  map[string]bool{},
    }
    return c
}

func (c *Hash) hashKey(key string) uint32 {
    return crc32.ChecksumIEEE([]byte(key))
}
func (c *Hash) Add(node string, vCounts ...int) bool {
    if node == "" {
        return false
    }
    c.Lock()
    defer c.Unlock()

    if _, ok := c.nodes[node]; ok {
        return false
    }
    vCount := 1
    if len(vCounts) == 1 {
        vCount = vCounts[0]
    }
    c.nodes[node] = true
    // 虚拟节点
    for i := 0; i < vCount; i++ {
        vKey := c.hashKey(node + strconv.Itoa(i))
        c.circle[vKey] = node
        c.sortedNodes = append(c.sortedNodes, vKey)
    }
    // 排序
    sort.Slice(c.sortedNodes, func(i, j int) bool {
        return c.sortedNodes[i] < c.sortedNodes[j]
    })
    return true
}
func (c *Hash) GetNode(node string) string {
    c.RLock()
    defer c.RUnlock()
    h := c.hashKey(node)
    i := c.getPosition(h)
    return c.circle[c.sortedNodes[i]]
}
func (c *Hash) Nodes() []string {
    c.RLock()
    defer c.RUnlock()
    var result []string
    for k, _ := range c.nodes {
        result = append(result, k)
    }
    return result
}
func (c *Hash) getPosition(h uint32) int {
    sz := len(c.sortedNodes)
    i := sort.Search(sz, func(i int) bool {
        return c.sortedNodes[i] > h
    })
    if i < sz {
        if i == sz-1 {
            return 0
        }
        return i
    }
    return sz - 1
}
