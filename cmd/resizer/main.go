package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"resizer/config"
	"resizer/logger"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)
	

func main() {
	var configFile string
	var versionFlag bool

	// Определение режима работы (по умолчанию production)
	isDevelopment := strings.ToLower(os.Getenv("ENV_APP")) == "dev"

	rootCmd := &cobra.Command{
		Use:   "resizer",
		Short: "Image resize service",
		Run: func(_ *cobra.Command, _ []string) {
			if versionFlag {
				printVersion()
				return
			}
			cfg, err := config.LoadConfig(configFile)
			if err != nil {
				log.Fatalf("failed to load config: %v", err)
			}

			logg, err := logger.NewLogger(cfg.Logger.Level, isDevelopment)
			if err != nil {
				log.Fatalf("failed to initialize logger: %v", err)
			}
			defer logg.Sync()

			var storage storagePackage.Storage

			storage = memorystorage.NewInMemoryStorage()

			logg.Info("calendar is running...")
			//calendar := app.New(logg, storage)
			//server := internalhttp.NewServer(
			//	logg,
			//	calendar,
			//	cfg.Server.Host,
			//	cfg.Server.Port,
			//	time.Duration(cfg.Server.ReadTimeout),
			//	time.Duration(cfg.Server.WriteTimeout),
			//)

			var server = config.CreateServer(
				os.Args[1:],
			)
			if server != nil {
				server.Run()
			}
		},
	}
	// Флаг для указания пути к конфигурационному файлу
	rootCmd.Flags().StringVar(&configFile, "config", "", "path to config file (required)")

	// Флаг --version
	rootCmd.Flags().BoolVar(&versionFlag, "version", false, "print the version of the application")
	// err := rootCmd.MarkFlagRequired("config")
	// if err != nil {
	//	return
	// }

	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("command execution failed: %v", err)
	}

}
