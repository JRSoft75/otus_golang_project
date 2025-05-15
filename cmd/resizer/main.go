package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv" //nolint:depguard
	"github.com/spf13/cobra"   //nolint:depguard
	"go.uber.org/zap"
	"resizer/config"         //nolint:depguard
	"resizer/internal/cache" //nolint:depguard
	"resizer/logger"         //nolint:depguard
)

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
		_ = logg.Sync()
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

			// Регистрация обработчиков
			http.HandleFunc("/resize/", ResizeHandler(lruCache, logg))

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
