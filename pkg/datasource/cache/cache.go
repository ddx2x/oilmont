package cache

import (
	"github.com/patrickmn/go-cache"
	"github.com/ddx2x/oilmont/pkg/datasource"
	"time"
)

var _ datasource.ICache = &Cache{}

type Cache struct {
	*cache.Cache
}

func (c *Cache) GetCache(k string) (interface{}, bool) {
	return c.Get(k)
}

func (c *Cache) SetCache(k string, x interface{}, d time.Duration) {
	c.Set(k, x, d)
}

func NewCache(defaultExpiration, cleanupInterval time.Duration) *Cache {
	c := &Cache{
		cache.New(defaultExpiration, cleanupInterval),
	}
	return c
}
