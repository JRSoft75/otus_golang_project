package cache

import (
	"os"
	"path/filepath"
	"sync"
)

type Cache interface {
	Set(key string, data []byte) error
	Get(key string) ([]byte, bool)
	Clear()
}

type cacheItem struct {
	key   string
	value interface{}
}

type lruCache struct {
	dir      string
	capacity int
	queue    List
	items    map[string]*ListItem
	mu       sync.Mutex
}

func NewCache(capacity int, dir string) (Cache, error) {
	// Создаем директорию для кэша, если её нет
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return nil, err
	}
	return &lruCache{
		dir:      dir,
		capacity: capacity,
		queue:    NewList(),
		items:    make(map[string]*ListItem, capacity),
	}, nil
}

func (c *lruCache) Set(key string, data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	filePath := filepath.Join(c.dir, key)
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return err
	}

	if item, found := c.items[key]; found {
		// Update the value and move to front
		item.Value.(*cacheItem).value = filePath
		c.queue.MoveToFront(item)
		return nil
	}

	newItem := &cacheItem{key: key, value: filePath}
	listItem := c.queue.PushFront(newItem)
	c.items[key] = listItem

	// Check capacity
	if c.queue.Len() > c.capacity {
		// Remove the least recently used item
		backItem := c.queue.Back()
		if backItem != nil {
			c.queue.Remove(backItem)
			delete(c.items, backItem.Value.(*cacheItem).key)
		}
	}

	return nil
}

func (c *lruCache) Get(key string) ([]byte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if item, found := c.items[key]; found {
		// Move to front and return value
		c.queue.MoveToFront(item)
		path := item.Value.(*cacheItem).value
		data, err := os.ReadFile(path.(string))
		if err == nil {
			return data, true
		}
	}
	return nil, false
}

func (c *lruCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.queue = NewList()
	c.items = make(map[string]*ListItem, c.capacity)
}
