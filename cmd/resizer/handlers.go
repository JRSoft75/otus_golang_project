package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"go.uber.org/zap"
	"resizer/internal/cache" //nolint:depguard
	"resizer/internal/image" //nolint:depguard
)

var slashRegex = regexp.MustCompile(`^/+`)

// ResizeHandler обрабатывает запросы на изменение размера изображений.
func ResizeHandler(lruCache cache.Cache, logg *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Удаляем префикс "/resize/"
		path := strings.TrimPrefix(r.URL.Path, "/resize/")
		// Разделяем путь на части по первым двум слешам
		parts := strings.SplitN(path, "/", 3)
		if len(parts) < 3 {
			http.Error(w, "Invalid URL format", http.StatusBadRequest)
			return
		}

		width, height, rawURL := parts[0], parts[1], parts[2]

		// Парсим URL для корректной обработки
		parsedURL, err := url.Parse(rawURL)
		if err != nil {
			http.Error(w, "Invalid URL format", http.StatusBadRequest)
			return
		}

		// Удаляем лишние слеши в начале URL
		parsedURL.Path = slashRegex.ReplaceAllString(parsedURL.Path, "")
		rawURL = "http://" + parsedURL.Path

		// Генерируем хэш от URL
		urlHash := GenerateHash(parsedURL.String())

		// Генерируем ключ для кэша
		cacheKey := fmt.Sprintf("%s_%s_%s", width, height, urlHash)

		// Проверяем наличие в кэше
		if data, ok := lruCache.Get(cacheKey); ok {
			format := cacheKey[strings.LastIndex(cacheKey, "_")+1:] // Извлекаем формат из ключа
			w.Header().Set("Content-Type", getContentType(format))
			_, err := w.Write(data)
			if err != nil {
				logg.Error(fmt.Sprintf("Failed to write response: %v", err))
			}
			return
		}

		// Загружаем и обрабатываем изображение
		data, err := image.DownloadImage(rawURL, r.Header) // Передаем заголовки исходного запроса
		if err != nil {
			http.Error(w, "Failed to download image", http.StatusInternalServerError)
			return
		}

		resizedData, format, err := image.ResizeImage(data, atoi(width), atoi(height))
		if err != nil {
			http.Error(w, "Failed to resize image", http.StatusInternalServerError)
			return
		}

		// Сохраняем в кэш
		if err := lruCache.Set(cacheKey, resizedData); err != nil {
			logg.Error(fmt.Sprintf("Failed to cache image: %v", err))
		}

		// Возвращаем изображение
		w.Header().Set("Content-Type", getContentType(format))
		_, err = w.Write(resizedData)
		if err != nil {
			logg.Error(fmt.Sprintf("Failed to write response: %v", err))
		}
	}
}

func atoi(s string) int {
	var n int
	_, err := fmt.Sscanf(s, "%d", &n)
	if err != nil {
		return 0
	}
	return n
}

func getContentType(format string) string {
	switch format {
	case "jpeg":
		return "image/jpeg"
	case "png":
		return "image/png"
	case "gif":
		return "image/gif"
	default:
		return "application/octet-stream"
	}
}

// GenerateHash создает SHA256 хэш от строки и возвращает его в виде шестнадцатеричной строки.
func GenerateHash(input string) string {
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])
}
