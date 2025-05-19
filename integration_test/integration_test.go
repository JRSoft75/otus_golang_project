package integration_test

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestResizeService(t *testing.T) {
	// Ждем, пока сервисы станут доступны
	time.Sleep(5 * time.Second)

	// Тест 1: Ресайз изображения image1.jpg
	t.Run("Resize image1.jpg", func(t *testing.T) {
		url := "http://localhost:8080/resize/300/200/http://resize-nginx/image1.jpg"
		// Создаем контекст с таймаутом 10 секунд
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Создаем новый HTTP-запрос с контекстом
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return
		}
		resp, err := http.DefaultClient.Do(req)
		// resp, err := http.Get(url)
		if err != nil {
			t.Fatalf("Failed to fetch resized image: %v", err)
		}
		// Гарантируем закрытие тела ответа
		defer func() {
			_ = resp.Body.Close()
		}()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		// Проверяем, что ответ содержит изображение
		contentType := resp.Header.Get("Content-Type")
		if !strings.HasPrefix(contentType, "image/") {
			t.Errorf("Expected image content type, got %s", contentType)
		}
	})

	// Тест 2: Ресайз изображения image3.png
	t.Run("Resize image3.png", func(t *testing.T) {
		url := "http://localhost:8080/resize/300/200/http://resize-nginx/image3.png"
		// Создаем контекст с таймаутом 10 секунд
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Создаем новый HTTP-запрос с контекстом
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return
		}
		resp, err := http.DefaultClient.Do(req)
		// resp, err := http.Get(url)
		if err != nil {
			t.Fatalf("Failed to fetch resized image: %v", err)
		}
		// Гарантируем закрытие тела ответа
		defer func() {
			_ = resp.Body.Close()
		}()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		// Проверяем, что ответ содержит изображение
		contentType := resp.Header.Get("Content-Type")
		if !strings.HasPrefix(contentType, "image/") {
			t.Errorf("Expected image content type, got %s", contentType)
		}
	})

	// Тест 3: Ресайз защищенного изображения image2.jpeg
	t.Run("Resize secure image2.jpeg with authorization", func(t *testing.T) {
		// Создаем контекст с таймаутом 10 секунд
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		url := "http://localhost:8080/resize/300/200/http://resize-nginx/secure/image2.jpeg"
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		req.Header.Set("Authorization", "Bearer your-token-here")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Failed to fetch resized image: %v", err)
		}
		// Гарантируем закрытие тела ответа
		defer func() {
			_ = resp.Body.Close()
		}()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		// Проверяем, что ответ содержит изображение
		contentType := resp.Header.Get("Content-Type")
		if !strings.HasPrefix(contentType, "image/") {
			t.Errorf("Expected image content type, got %s", contentType)
		}
	})
}
