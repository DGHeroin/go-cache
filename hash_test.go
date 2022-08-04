package cache

import "testing"

func TestNewHash(t *testing.T) {
    h := NewHash()
    h.Add("a", 10)
    h.Add("b", 10)
    h.Add("c", 10)

    t.Log(h.GetNode("d"))
}
