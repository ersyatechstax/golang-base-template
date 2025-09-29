package cache

import (
	"context"
	"sync"

	"github.com/allegro/bigcache/v3"
	"github.com/pkg/errors"
)

type IBigCache interface {
	Get(key string) ([]byte, error)
	Set(key string, data []byte) error
}
type bigCache struct {
	bigCache *bigcache.BigCache
	mx       *sync.Mutex
}

func NewBigCache() IBigCache {
	// TODO: Set default config

	bc, err := bigcache.New(context.Background(), bigcache.Config{
		Shards:             32,       // must be the power of 2
		LifeWindow:         3600,     // ttl
		CleanWindow:        360,      // 1/10 life-window
		MaxEntriesInWindow: 15000000, // max entries in life-window, 15M
		MaxEntrySize:       8,        // in bytes, an int64 is 8 bytes
	})

	if err != nil {
		return nil
	}

	return &bigCache{
		bigCache: bc,
		mx:       new(sync.Mutex),
	}
}

func (bc *bigCache) Get(key string) ([]byte, error) {
	if bc == nil || bc.bigCache == nil {
		return nil, errors.New("bigcache is nil")
	}
	bc.mx.Lock()
	defer bc.mx.Unlock()
	return bc.bigCache.Get(key)
}

func (bc *bigCache) Set(key string, data []byte) error {
	if bc == nil || bc.bigCache == nil {
		return errors.New("bigcache is nil")
	}
	bc.mx.Lock()
	defer bc.mx.Unlock()
	return bc.bigCache.Set(key, data)
}
