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
	// Создаем путь к файлу на основе хэша
	filePath := filepath.Join(c.dir, key)
	if err := os.WriteFile(filePath, data, 0o644); err != nil {
		return err
	}

	if item, found := c.items[key]; found {
		// обновляем и перемещаем вперед списка
		item.Value.(*cacheItem).value = filePath
		c.queue.MoveToFront(item)
		return nil
	}

	newItem := &cacheItem{key: key, value: filePath}
	listItem := c.queue.PushFront(newItem)
	c.items[key] = listItem

	if c.queue.Len() > c.capacity {
		// удаляем давно не использованный элемент
		backItem := c.queue.Back()
		if backItem != nil {
			path := backItem.Value.(*cacheItem).value
			_ = os.Remove(path.(string))
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
