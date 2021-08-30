package squirrel

import (
	"sync"
)

type Cache struct {
	stashes map[interface{}]*Stash
	lock sync.RWMutex

	Find func(key interface{}) interface{}
}

func NewCache() *Cache {
	return &Cache{stashes: make(map[interface{}]*Stash)}
}

func (c *Cache) UpsertValue(key interface{}, val interface{}) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.stashes[key] = NewStash(val).Now()
}

func (c *Cache) UpsertStash(key interface{}, s *Stash) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.stashes[key] = s
}

func (c *Cache) GetStash(key interface{}) *Stash {
	// Don't use a defer here, since c.UpsertValue might need the lock before the end of the func
	c.lock.RLock()

	// Try to Find it in direct cache
	stash, found := c.stashes[key]
	if found {
		c.lock.RUnlock()
		return stash
	}

	c.lock.RUnlock()

	// Not in the cache. If Find is set we use it to search for the value
	if c.Find != nil {
		res := c.Find(key)
		if res != nil {
			stash := NewStash(res).Now()
			c.UpsertStash(key, stash)

			return stash
		}
	}

	// Not found anywhere
	return nil
}

func (c *Cache) Get(key interface{}) interface{} {
	res := c.GetStash(key)
	if res != nil {
		return res.val
	}

	return nil
}

func (c *Cache) SearchStash(searchFunc func(interface{}) bool) []*Stash {
	c.lock.RLock()
	defer c.lock.RUnlock()

	var results []*Stash
	for _, s := range c.stashes {
		if searchFunc(s.val) {
			results = append(results, s)
		}
	}

	return results
}

func (c *Cache) Search(searchFunc func(interface{}) bool) []interface{} {
	results := c.SearchStash(searchFunc)

	var resultsVals []interface{}
	for _, s := range results {
		resultsVals = append(resultsVals, s.val)
	}

	return resultsVals
}

func (c *Cache) UpdateIfNewer(key interface{}, s *Stash) {
	// Don't use a defer here, since c.UpsertStash might need the lock before the end of the func
	c.lock.RLock()

	// We avoid the Get function since we don't want to fall back to the c.Find call
	current, found := c.stashes[key]

	c.lock.RUnlock()

	if !found {
		// No previous value. Just add it
		c.UpsertStash(key, s)
	}

	if s.status.creation.After(current.status.creation) {
		// The new value is newer. Upsert it
		c.UpsertStash(key, s)
	}

	// The new value is newer. Keep it
}

func (c *Cache) Delete(key interface{}) {
	c.lock.Lock()
	defer c.lock.Unlock()

	delete(c.stashes, key)
}
