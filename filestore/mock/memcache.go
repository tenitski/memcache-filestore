package mock

import (
	"filestore/client"

	"github.com/bradfitz/gomemcache/memcache"
)

type mockMemcacheClient struct {
	store       map[string][]byte
	maxCapacity int
}

func NewMemcacheClient(maxCapacity int) client.Memcache {
	return &mockMemcacheClient{
		store:       map[string][]byte{},
		maxCapacity: maxCapacity,
	}
}

func (c *mockMemcacheClient) Set(item *memcache.Item) error {
	if len(c.store) > c.maxCapacity {
		// No more room in the store
		return nil
	}

	c.store[item.Key] = item.Value
	return nil
}

func (c *mockMemcacheClient) Get(key string) (item *memcache.Item, err error) {
	if value, ok := c.store[key]; ok {
		return &memcache.Item{Key: key, Value: value}, nil
	}

	return nil, memcache.ErrCacheMiss
}

func (c *mockMemcacheClient) GetMulti(keys []string) (map[string]*memcache.Item, error) {
	items := map[string]*memcache.Item{}
	for _, key := range keys {
		if value, ok := c.store[key]; ok {
			items[key] = &memcache.Item{Key: key, Value: value}
		}
	}

	return items, nil
}

func (c *mockMemcacheClient) Delete(key string) error {
	if _, ok := c.store[key]; ok {
		delete(c.store, key)
		return nil
	}

	return memcache.ErrCacheMiss
}
