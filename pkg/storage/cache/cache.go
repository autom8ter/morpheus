package cache

import (
	"context"
	"sync"
	"time"
)

type Cache struct {
	mu         sync.RWMutex
	records    map[string]*cachedItem
	maxRecords int
	gcInterval time.Duration
}

func NewCache(maxRecords int, gcInterval time.Duration) *Cache {
	return &Cache{
		mu:         sync.RWMutex{},
		records:    map[string]*cachedItem{},
		maxRecords: maxRecords,
		gcInterval: gcInterval,
	}
}

type cachedItem struct {
	values           map[string]interface{}
	last_interaction time.Time
	commited         bool
}

func (c *Cache) Get(key string) (map[string]interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	item := c.records[key]
	if item == nil {
		return nil, false
	}
	return item.values, true
}

func (c *Cache) Set(key string, values map[string]interface{}, isCommited bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	item := c.records[key]
	if item == nil {
		c.records[key] = &cachedItem{
			values:           values,
			last_interaction: time.Now(),
			commited:         isCommited,
		}
		return
	}
	item.values = values
	item.last_interaction = time.Now()
	item.commited = isCommited
}

func (c *Cache) GC(ctx context.Context) {

	if c.gcInterval == 0 {
		c.gcInterval = 1 * time.Second
	}
	tick := time.NewTicker(c.gcInterval)
	if c.maxRecords == 0 {
		c.maxRecords = 100000
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			if ctx.Err() != nil {
				return
			}
			c.mu.Lock()
			defer c.mu.Unlock()
			if total := len(c.records); total > c.maxRecords {
				deleted := 0
				for key, item := range c.records {
					if total-deleted <= c.maxRecords {
						break
					}
					if !item.commited {
						continue
					}
					delete(c.records, key)
					deleted++
				}
			}
		}
	}
}
