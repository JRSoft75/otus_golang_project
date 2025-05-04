package config

import (
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"flag"
	"fmt"
	"os"
	"resizer"
	"runtime"
	"strings"
	"time"

	"github.com/peterbourgon/ff/v3"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/yaml.v3"
	"resizer/server"
)

var baseConfig = []Option{
	withFileSystem,
	withHTTPLoader,
}

// LoggerConfig представляет настройки логгера.
type LoggerConfig struct {
	Level string `yaml:"level"`
}

type ServerConfig struct {
	Host        string `yaml:"host"`
	Port        int    `yaml:"port"`
	ReadTimeout int    `yaml:"readTimeout"`
}

// Config представляет основную структуру конфигурации сервиса.
type Config struct {
	Logger  LoggerConfig `yaml:"logger"`
	Server  ServerConfig `yaml:"server"`
	Storage struct {
		FileCount int `yaml:"file_count"`
	} `yaml:"storage"`
}

func LoadConfig(filePath string) (*Config, error) {
	// Проверка существования файла
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file does not exist: %s", filePath)
	}

	// Чтение содержимого файла
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Парсинг YAML
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

// NewImagor create imagor from config flags
//func NewResizer(
//	fs *flag.FlagSet, cb func() (*zap.Logger, bool), funcs ...Option,
//) *resizer.Resizer {
//	var (
//		imagorRequestTimeout = fs.Duration("imagor-request-timeout",
//			time.Second*30, "Timeout for performing imagor request")
//		imagorLoadTimeout = fs.Duration("imagor-load-timeout",
//			0, "Timeout for imagor Loader request, should be smaller than imagor-request-timeout")
//		imagorSaveTimeout = fs.Duration("imagor-save-timeout",
//			0, "Timeout for saving image to imagor Storage")
//		imagorProcessTimeout = fs.Duration("imagor-process-timeout",
//			0, "Timeout for image processing")
//		imagorBaseParams = fs.String("imagor-base-params", "",
//			"imagor endpoint base params that applies to all resulting images e.g. filters:watermark(example.jpg)")
//		imagorProcessConcurrency = fs.Int64("imagor-process-concurrency",
//			-1, "Maximum number of image process to be executed simultaneously. Requests that exceed this limit are put in the queue. Set -1 for no limit")
//		imagorProcessQueueSize = fs.Int64("imagor-process-queue-size",
//			0, "Maximum number of image process that can be put in the queue. Requests that exceed this limit are rejected with HTTP status 429")
//
//		imagorDisableParamsEndpoint  = fs.Bool("imagor-disable-params-endpoint", false, "imagor disable /params endpoint")
//		imagorSignerType             = fs.String("imagor-signer-type", "sha1", "imagor URL signature hasher type: sha1, sha256, sha512")
//		imagorStoragePathStyle       = fs.String("imagor-storage-path-style", "original", "imagor storage path style: original, digest")
//		imagorResultStoragePathStyle = fs.String("imagor-result-storage-path-style", "original", "imagor result storage path style: original, digest, suffix")
//
//		options, logger, isDebug = applyOptions(fs, cb, append(funcs, baseConfig...)...)
//
//		alg          = sha1.New
//		hasher       imagorpath.StorageHasher
//		resultHasher imagorpath.ResultStorageHasher
//	)
//
//	if strings.ToLower(*imagorSignerType) == "sha256" {
//		alg = sha256.New
//	} else if strings.ToLower(*imagorSignerType) == "sha512" {
//		alg = sha512.New
//	}
//
//	if strings.ToLower(*imagorStoragePathStyle) == "digest" {
//		hasher = imagorpath.DigestStorageHasher
//	}
//
//	if strings.ToLower(*imagorResultStoragePathStyle) == "digest" {
//		resultHasher = imagorpath.DigestResultStorageHasher
//	} else if strings.ToLower(*imagorResultStoragePathStyle) == "suffix" {
//		resultHasher = imagorpath.SuffixResultStorageHasher
//	} else if strings.ToLower(*imagorResultStoragePathStyle) == "size" {
//		resultHasher = imagorpath.SizeSuffixResultStorageHasher
//	}
//
//	return imagor.New(append(
//		options,
//		imagor.WithBasePathRedirect(*imagorBasePathRedirect),
//		imagor.WithBaseParams(*imagorBaseParams),
//		imagor.WithRequestTimeout(*imagorRequestTimeout),
//		imagor.WithLoadTimeout(*imagorLoadTimeout),
//		imagor.WithSaveTimeout(*imagorSaveTimeout),
//		imagor.WithProcessTimeout(*imagorProcessTimeout),
//		imagor.WithProcessConcurrency(*imagorProcessConcurrency),
//		imagor.WithProcessQueueSize(*imagorProcessQueueSize),
//		imagor.WithCacheHeaderTTL(*imagorCacheHeaderTTL),
//		imagor.WithCacheHeaderSWR(*imagorCacheHeaderSWR),
//		imagor.WithDisableParamsEndpoint(*imagorDisableParamsEndpoint),
//		imagor.WithStoragePathStyle(hasher),
//		imagor.WithResultStoragePathStyle(resultHasher),
//		imagor.WithLogger(logger),
//		imagor.WithDebug(isDebug),
//	)...)
//}

// CreateServer create server from config flags. Returns nil on version or help command
func CreateServer(args []string, funcs ...Option) (srv *server.Server) {
	var (
		fs     = flag.NewFlagSet("imagor", flag.ExitOnError)
		logger *zap.Logger
		err    error
		app    *resizer.Resizer

		debug        = fs.Bool("debug", false, "Debug mode")
		version      = fs.Bool("version", false, "imagor version")
		port         = fs.Int("port", 8000, "Server port")
		goMaxProcess = fs.Int("gomaxprocs", 0, "GOMAXPROCS")

		bind = fs.String("bind", "",
			"Server address and port to bind .e.g. myhost:8888. This overrides server address and port config")

		_ = fs.String("config", ".env", "Retrieve configuration from the given file")

		serverAddress = fs.String("server-address", "",
			"Server address")
		serverPathPrefix = fs.String("server-path-prefix", "",
			"Server path prefix")
		serverCORS = fs.Bool("server-cors", false,
			"Enable CORS")
		serverStripQueryString = fs.Bool("server-strip-query-string", false,
			"Enable strip query string redirection")
		serverAccessLog = fs.Bool("server-access-log", false,
			"Enable server access log")
	)

	app = NewImagor(fs, func() (*zap.Logger, bool) {
		if err = ff.Parse(fs, args,
			ff.WithEnvVars(),
			ff.WithConfigFileFlag("config"),
			ff.WithIgnoreUndefined(true),
			ff.WithAllowMissingConfigFile(true),
			ff.WithConfigFileParser(ff.EnvParser),
		); err != nil {
			panic(err)
		}
		if *debug {
			logger = zap.Must(zap.NewDevelopment())
		} else {
			logger = zap.Must(zap.NewProduction())
		}

		return logger, *debug
	}, funcs...)

	if *goMaxProcess > 0 {
		logger.Debug("GOMAXPROCS", zap.Int("count", *goMaxProcess))
		runtime.GOMAXPROCS(*goMaxProcess)
	}

	return server.New(app,
		server.WithAddr(*bind),
		server.WithPort(*port),
		server.WithAddress(*serverAddress),
		server.WithPathPrefix(*serverPathPrefix),
		server.WithCORS(*serverCORS),
		server.WithStripQueryString(*serverStripQueryString),
		server.WithAccessLog(*serverAccessLog),
		server.WithLogger(logger),
		server.WithDebug(*debug),
	)
}
