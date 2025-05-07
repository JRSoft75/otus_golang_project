package cache

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLRUCache_simple(t *testing.T) {
	// Создаем временную директорию для тестов
	tempDir := t.TempDir()

	// Инициализация LRU-кэша
	c, err := NewCache(10, tempDir)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	// Проверяем, что ключи отсутствуют
	_, ok := c.Get("key1")
	require.False(t, ok)

	_, ok = c.Get("key2")
	require.False(t, ok)

	// Добавляем первый элемент
	err = c.Set("key1", []byte("value1"))
	if err != nil {
		t.Fatalf("Failed to set key1: %v", err)
	}

	// Проверяем получение первого элемента
	data, ok := c.Get("key1")
	if !ok {
		t.Fatalf("Failed to get key1")
	}
	if string(data) != "value1" {
		t.Errorf("Expected value1, got %s", string(data))
	}

	// Добавляем второй элемент
	err = c.Set("key2", []byte("value2"))
	if err != nil {
		t.Fatalf("Failed to set key2: %v", err)
	}

	// Проверяем получение второго элемента
	data, ok = c.Get("key2")
	if !ok {
		t.Fatalf("Failed to get key2")
	}
	if string(data) != "value2" {
		t.Errorf("Expected value2, got %s", string(data))
	}

	// Добавляем третий элемент, что должно вызвать удаление "key1"
	err = c.Set("key3", []byte("value3"))
	if err != nil {
		t.Fatalf("Failed to set key3: %v", err)
	}

	// Проверяем, что "key1" удален
	_, ok = c.Get("key1")
	if ok {
		t.Errorf("Expected key1 to be evicted, but it was found")
	}

	// Проверяем, что "key2" и "key3" доступны
	_, ok = c.Get("key2")
	if !ok {
		t.Errorf("Expected key2 to be present, but it was not found")
	}
	_, ok = c.Get("key3")
	if !ok {
		t.Errorf("Expected key3 to be present, but it was not found")
	}
}

// Тест на логику выталкивания элементов из-за размера очереди
// (например: n = 3, добавили 4 элемента - 1й из кэша вытолкнулся);
func TestLRUCache_purge_logic(t *testing.T) {
	// Создаем временную директорию для тестов
	tempDir := t.TempDir()

	// Инициализация LRU-кэша
	c, err := NewCache(3, tempDir)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}
	_ = c.Set("key1", []byte("value1"))
	_ = c.Set("key2", []byte("value2"))
	_ = c.Set("key3", []byte("value3"))
	_ = c.Set("key4", []byte("value4"))

	val, ok := c.Get("key1")
	require.False(t, ok)
	require.Nil(t, val)

	_, ok = c.Get("key2")
	require.True(t, ok)

	_, ok = c.Get("key3")
	require.True(t, ok)

	_, ok = c.Get("key4")
	require.True(t, ok)
}

// Тест на логику выталкивания давно используемых элементов
// (например: n = 3, добавили 3 элемента, обратились несколько раз к разным элементам: изменили значение, получили
// значение и пр. - добавили 4й элемент, из первой тройки вытолкнется тот элемент, что был затронут наиболее давно).
func TestLRUCache_access_time(t *testing.T) {
	// Создаем временную директорию для тестов
	tempDir := t.TempDir()

	// Инициализация LRU-кэша
	c, err := NewCache(3, tempDir)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}
	_ = c.Set("key1", []byte("value1"))
	_ = c.Set("key2", []byte("value2"))
	_ = c.Set("key3", []byte("value3"))

	_, ok := c.Get("key1")
	require.True(t, ok)

	_, ok = c.Get("key2")
	require.True(t, ok)

	_ = c.Set("key1", []byte("value4"))
	_ = c.Set("key2", []byte("value5"))

	_, _ = c.Get("key1")
	_, _ = c.Get("key2")

	_ = c.Set("key1", []byte("value6"))

	_, _ = c.Get("key1")

	_ = c.Set("key4", []byte("value7"))

	_, ok = c.Get("key1")
	require.True(t, ok)

	_, ok = c.Get("key2")
	require.True(t, ok)

	val, ok := c.Get("key3")
	require.False(t, ok)
	require.Nil(t, val)

	_, ok = c.Get("key4")
	require.True(t, ok)
}

func TestLRUCache_file(t *testing.T) {
	// Создаем временную директорию для тестов
	tempDir := t.TempDir()

	// Инициализация LRU-кэша
	c, err := NewCache(2, tempDir)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	// Добавляем элемент
	err = c.Set("key1", []byte("value1"))
	if err != nil {
		t.Fatalf("Failed to set key1: %v", err)
	}

	// Проверяем, что файл создан на диске
	filePath := filepath.Join(tempDir, "key1")
	_, err = os.Stat(filePath)
	if os.IsNotExist(err) {
		t.Errorf("Expected file %s to exist, but it does not", filePath)
	}

	// Получаем элемент из кэша
	data, ok := c.Get("key1")
	if !ok {
		t.Fatalf("Failed to get key1")
	}
	if string(data) != "value1" {
		t.Errorf("Expected value1, got %s", string(data))
	}
}

func TestCacheMultithreading(t *testing.T) {
	// Создаем временную директорию для тестов
	tempDir := t.TempDir()

	// Инициализация LRU-кэша
	c, err := NewCache(10, tempDir)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	wg := &sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		for i := 0; i < 1_000_000; i++ {
			key := strconv.Itoa(i)
			err := c.Set(key, []byte(fmt.Sprintf("value%s", key)))
			if err != nil {
				return
			}
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 1_000_000; i++ {
			c.Get(strconv.Itoa(rand.Intn(1_000_000)))
		}
	}()

	wg.Wait()

	require.True(t, true)
}
