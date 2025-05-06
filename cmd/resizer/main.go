package main

import (
	//"context"
	//"flag"
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

	//"os/signal"
	//"resizer"
	"resizer/config"
	"resizer/internal/cache"
	"resizer/internal/image"
	"resizer/logger"
	//"resizer/server"
	//"strconv"
	"strings"
)

var slashRegex = regexp.MustCompile(`^/+`)

//func getIntEnv(key string, defaultValue int) int {
//	if value, exists := os.LookupEnv(key); exists {
//		var intValue int
//		_, err := fmt.Sscanf(value, "%d", &intValue)
//		if err == nil {
//			return intValue
//		}
//	}
//	return defaultValue
//}
//
//func getStringEnv(key, defaultValue string) string {
//	if value, exists := os.LookupEnv(key); exists {
//		return value
//	}
//	return defaultValue
//}

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

				// Генерируем ключ для кэша
				cacheKey := fmt.Sprintf("%s_%s_%s", width, height, rawURL)

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
				if err := lruCache.Set(cacheKey+"_"+format, resizedData); err != nil {
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

//func runServer(ctx context.Context) error {
//	//if err := checkVipsVersion(bimg.VipsMajorVersion, bimg.VipsMinorVersion); err != nil {
//	//	return err
//	//}
//	//configPath := flag.String("config", "", "Path of config file in yml format")
//	//flag.Parse()
//	//if *configPath == "" {
//	//	return fmt.Errorf("Set config.yml path via -config flag.")
//	//}
//	//file, err := os.Open(*configPath)
//	//if err != nil {
//	//	return fmt.Errorf("Error loading config: %v", err)
//	//}
//	//config, err := parseConfig(file)
//	//file.Close()
//	//if err != nil {
//	//	return err
//	//}
//	//if config.LogPath != "" {
//	//	logFile, err := os.OpenFile(config.LogPath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
//	//	if err != nil {
//	//		return fmt.Errorf("Could not open log file: %v", err)
//	//	}
//	//	defer logFile.Close()
//	//	log.SetOutput(logFile)
//	//} else {
//	//	log.SetOutput(os.Stdout)
//	//}
//
//	//server := createServer(config)
//	//
//	//done := make(chan os.Signal, 1)
//	//signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
//	//defer close(done)
//	//
//	//serverErr := make(chan error)
//	//defer close(serverErr)
//	//
//	//go func() {
//	//	log.Printf("Starting server on %s", config.ServerAddress)
//	//	if err := server.ListenAndServe(config.ServerAddress); err != nil {
//	//		serverErr <- err
//	//	}
//	//}()
//	//
//	//select {
//	//case <-done:
//	//	return server.Shutdown()
//	//case <-ctx.Done():
//	//	return server.Shutdown()
//	//case err := <-serverErr:
//	//	return err
//	//}
//}
//
////func main() {
////	ctx := context.Background()
////	err := runServer(ctx)
////	if err != nil {
////		log.Fatal(err)
////		ctx.Done()
////	}
////
////}
//
//func main() {
//	var configFile string
//	var versionFlag bool
//
//	// Определение режима работы (по умолчанию production)
//	isDevelopment := strings.ToLower(os.Getenv("ENV_APP")) == "dev"
//
//	rootCmd := &cobra.Command{
//		Use:   "resizer",
//		Short: "Image resize service",
//		Run: func(_ *cobra.Command, _ []string) {
//			if versionFlag {
//				printVersion()
//				return
//			}
//			cfg, err := config.LoadConfig(configFile)
//			if err != nil {
//				log.Fatalf("failed to load config: %v", err)
//			}
//
//			logg, err := logger.NewLogger(cfg.Logger.Level, isDevelopment)
//			if err != nil {
//				log.Fatalf("failed to initialize logger: %v", err)
//			}
//			defer func(logg *zap.Logger) {
//				err := logg.Sync()
//				if err != nil {
//
//				}
//			}(logg)
//
//			//var storage storagePackage.Storage
//			//
//			//storage = memorystorage.NewInMemoryStorage()
//
//			logg.Info("Storage is running...")
//			//calendar := app.New(logg, storage)
//			serv := server.NewServer(
//				logg,
//				//calendar,
//				cfg.Server.Host,
//				cfg.Server.Port,
//				//time.Duration(cfg.Server.ReadTimeout),
//				//time.Duration(cfg.Server.WriteTimeout),
//			)
//			ctx, cancel := signal.NotifyContext(context.Background(),
//				syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
//			defer cancel()
//
//			go func() {
//				<-ctx.Done()
//
//				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
//				defer cancel()
//
//				if err := serv.Stop(ctx); err != nil {
//					logg.Error("failed to stop http server: " + err.Error())
//				}
//			}()
//
//			if err := serv.Start(ctx); err != nil {
//				logg.Error("failed to start http server: " + err.Error())
//				cancel()
//				os.Exit(1)
//			}
//			//var serv = serv.NewServer(
//			//	resizer.NewApp(),
//			//	cfg.Server.Host,
//			//	cfg.Server.Port,
//			//	time.Duration(cfg.Server.ReadTimeout),
//			//	logg,
//			//	os.Args[1:],
//			//)
//			//if serv != nil {
//			//	serv.Run()
//			//}
//		},
//	}
//	// Флаг для указания пути к конфигурационному файлу
//	rootCmd.Flags().StringVar(&configFile, "config", "", "path to config file (required)")
//
//	// Флаг --version
//	rootCmd.Flags().BoolVar(&versionFlag, "version", false, "print the version of the application")
//	// err := rootCmd.MarkFlagRequired("config")
//	// if err != nil {
//	//	return
//	// }
//
//	if err := rootCmd.Execute(); err != nil {
//		log.Fatalf("command execution failed: %v", err)
//	}
//
//}
