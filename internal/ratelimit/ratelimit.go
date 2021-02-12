package ratelimit

import (
	"errors"
	"sync"
)

var RateLimitError = errors.New("rate limit")

var rate = rateLimit{
	buckets: map[string]*bucket{},
}

type rateLimit struct {
	sync.RWMutex

	buckets map[string]*bucket
}

type bucket struct {
	sync.RWMutex

	key  string
	data int64
}

func Add(key string, max int64) error {
	rate.Lock()
	defer rate.Unlock()

	bkt, ok := rate.buckets[key]
	if !ok {
		if max <= 0 {
			return RateLimitError
		}
		rate.buckets[key] = &bucket{
			RWMutex: sync.RWMutex{},
			key:     key,
			data:    1,
		}
		return nil
	}

	bkt.Lock()
	defer bkt.Unlock()
	if bkt.data+1 > max {
		return RateLimitError
	}
	bkt.data++
	return nil
}

func Get(key string) int64 {
	rate.RLock()
	defer rate.Unlock()

	bkt, ok := rate.buckets[key]
	if !ok {
		return 0
	}

	bkt.RLock()
	defer bkt.RUnlock()
	return bkt.data
}

func Done(key string) {
	rate.Lock()
	defer rate.Unlock()

	bkt, ok := rate.buckets[key]
	if !ok {
		return
	}

	bkt.Lock()
	defer bkt.Unlock()
	bkt.data--
	if bkt.data < 0 {
		bkt.data = 0
	}
}
