package core

import (
	"sync"
	"time"
)

type CacheActive struct {
	Mutex *sync.Mutex
	Map   map[string]int64
}

func (c *CacheActive) clean() {
	t := time.Now().UnixMilli() - (time.Second * 30).Milliseconds()
	for activeId, ts := range c.Map {
		if ts <= t {
			delete(c.Map, activeId)
		}
	}
}

func (c *CacheActive) Add(activeId string) bool {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	c.clean()
	_, ok := c.Map[activeId]
	if !ok {
		c.Map[activeId] = time.Now().UnixMilli()
	}
	return ok
}
