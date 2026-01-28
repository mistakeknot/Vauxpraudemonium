package server

import (
	"container/list"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sync"
	"time"
)

type cacheEntry struct {
	key     string
	value   any
	expires time.Time
}

type inflight struct {
	done  chan struct{}
	value any
	err   error
}

// ScanCache provides TTL + LRU caching with in-flight de-duplication.
type ScanCache struct {
	mu       sync.Mutex
	items    map[string]*list.Element
	order    *list.List
	inflight map[string]*inflight
	max      int
}

func NewScanCache(max int) *ScanCache {
	return &ScanCache{
		items:    make(map[string]*list.Element),
		order:    list.New(),
		inflight: make(map[string]*inflight),
		max:      max,
	}
}

func (c *ScanCache) Get(key string) (any, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	el, ok := c.items[key]
	if !ok {
		return nil, false
	}
	entry := el.Value.(*cacheEntry)
	if time.Now().After(entry.expires) {
		c.order.Remove(el)
		delete(c.items, key)
		return nil, false
	}
	c.order.MoveToFront(el)
	return entry.value, true
}

func (c *ScanCache) GetOrCompute(key string, ttl time.Duration, fn func() (any, error)) (any, error) {
	if val, ok := c.Get(key); ok {
		return val, nil
	}

	c.mu.Lock()
	if in, ok := c.inflight[key]; ok {
		c.mu.Unlock()
		<-in.done
		return in.value, in.err
	}
	in := &inflight{done: make(chan struct{})}
	c.inflight[key] = in
	c.mu.Unlock()

	val, err := fn()

	c.mu.Lock()
	if err == nil {
		c.setLocked(key, val, ttl)
	}
	in.value = val
	in.err = err
	close(in.done)
	delete(c.inflight, key)
	c.mu.Unlock()

	return val, err
}

func (c *ScanCache) setLocked(key string, value any, ttl time.Duration) {
	if el, ok := c.items[key]; ok {
		entry := el.Value.(*cacheEntry)
		entry.value = value
		entry.expires = time.Now().Add(ttl)
		c.order.MoveToFront(el)
		return
	}
	entry := &cacheEntry{key: key, value: value, expires: time.Now().Add(ttl)}
	el := c.order.PushFront(entry)
	c.items[key] = el
	for c.max > 0 && c.order.Len() > c.max {
		back := c.order.Back()
		if back == nil {
			break
		}
		entry := back.Value.(*cacheEntry)
		delete(c.items, entry.key)
		c.order.Remove(back)
	}
}

func hashKey(v any) string {
	buf, _ := json.Marshal(v)
	sum := sha256.Sum256(buf)
	return hex.EncodeToString(sum[:])
}
