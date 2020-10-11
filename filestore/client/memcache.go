package client

import "github.com/bradfitz/gomemcache/memcache"

type Memcache interface {
	Set(item *memcache.Item) error
	Get(key string) (item *memcache.Item, err error)
	GetMulti(keys []string) (map[string]*memcache.Item, error)
	Delete(key string) error
}
