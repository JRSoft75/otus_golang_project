package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"

	"resizer/config"
	"resizer/internal/cache"
	"resizer/internal/image"
	"resizer/logger"
	"strings"

	"crypto/sha256"
	"encoding/hex"
)

var slashRegex = regexp.MustCompile(`^/+`)

func main() {
	var versionFlag bool
	configFile := "./config/config.yaml"
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Загрузка переменных окружения из файла .env
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables or defaults")
	}

	// Определение режима работы (по умолчанию production)
	isDevelopment := strings.ToLower(os.Getenv("ENV_APP")) == "dev"

	logg, err := logger.NewLogger(cfg.Logger.Level, isDevelopment)
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}
	defer func(logg *zap.Logger) {
		err := logg.Sync()
		if err != nil {

		}
	}(logg)

	rootCmd := &cobra.Command{
		Use:   "resizer",
		Short: "Image resize service",
		Run: func(_ *cobra.Command, _ []string) {
			if versionFlag {
				printVersion()
				return
			}

			logg.Info("Storage is running...")
			// Инициализация LRU-кэша
			lruCache, err := cache.NewCache(cfg.Storage.CacheSize, cfg.Storage.CacheDir)
			if err != nil {
				logg.Error(fmt.Sprintf("Failed to initialize cache: %v", err))
				return
			}

			http.HandleFunc("/resize/", func(w http.ResponseWriter, r *http.Request) {
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
						return
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
					return
				}
			})
			logg.Info(fmt.Sprintf("Starting server on : %s...", strconv.Itoa(cfg.Server.Port)))
			err = http.ListenAndServe(fmt.Sprintf(":%s", strconv.Itoa(cfg.Server.Port)), nil)
			if err != nil {
				return
			}
		},
	}

	// Флаг --version
	rootCmd.Flags().BoolVar(&versionFlag, "version", false, "print the version of the application")

	if err := rootCmd.Execute(); err != nil {
		logg.Fatal(fmt.Sprintf("command execution failed: %v", err))
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
