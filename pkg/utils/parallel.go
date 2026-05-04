package utils

import (
	"sync"
)

type WorkerPool struct {
	workers   int
	taskQueue chan func()
	wg        sync.WaitGroup
}

func NewWorkerPool(workers int) *WorkerPool {
	pool := &WorkerPool{
		workers:   workers,
		taskQueue: make(chan func(), 100),
	}

	pool.start()
	return pool
}

func (p *WorkerPool) start() {
	for i := 0; i < p.workers; i++ {
		go p.worker()
	}
}

func (p *WorkerPool) worker() {
	for task := range p.taskQueue {
		task()
		p.wg.Done()
	}
}

func (p *WorkerPool) Submit(task func()) {
	p.wg.Add(1)
	p.taskQueue <- task
}

func (p *WorkerPool) Wait() {
	p.wg.Wait()
}

func (p *WorkerPool) Close() {
	close(p.taskQueue)
}

func ParallelProcess(items []interface{}, processor func(interface{}) error, workers int) []error {
	var mu sync.Mutex
	errors := make([]error, 0)

	pool := NewWorkerPool(workers)
	defer pool.Close()

	for _, item := range items {
		item := item
		pool.Submit(func() {
			if err := processor(item); err != nil {
				mu.Lock()
				errors = append(errors, err)
				mu.Unlock()
			}
		})
	}

	pool.Wait()
	return errors
}

type CacheItem struct {
	Value      interface{}
	Expiration int64
}

type Cache struct {
	items map[string]*CacheItem
	mu    sync.RWMutex
}

func NewCache() *Cache {
	return &Cache{
		items: make(map[string]*CacheItem),
	}
}

func (c *Cache) Set(key string, value interface{}, ttl int64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = &CacheItem{
		Value:      value,
		Expiration: ttl,
	}
}

func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.items[key]
	if !exists {
		return nil, false
	}

	return item.Value, true
}

func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)
}

func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]*CacheItem)
}

func (c *Cache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.items)
}
