package config

import (
	"errors"
	"flag"
	"log/slog"
	"sync"
)

var (
	ErrInvalidLogLevel = errors.New("config: invalid log level, must be [-4, 0, 4, 8]")
	ErrInvalidPort     = errors.New("config: invalid port, must be between 1 and 65535")
)

var (
	cfg  *config
	once sync.Once
)

const (
	defaultPort     = 8080
	defaultLogLevel = int(slog.LevelInfo)
	defaultSecurity = true
)

func GetConfig() *config {
	once.Do(func() {
		cfg = initConfig()
	})
	return cfg
}

func initConfig() *config {
	port := flag.Int("p", defaultPort, "Server port")
	logLevel := flag.Int("log", int(defaultLogLevel), "Log Level [-4,0,4,8]")
	ssl := flag.Bool("s", true, "Secured connection (true/false)")
	flag.Parse()

	if err := validatePort(*port); err != nil {
		*port = defaultPort
		slog.Error("Validate port:", "error", err, "default value", defaultPort)
	}

	if err := validateLogLevel(*logLevel); err != nil {
		*logLevel = defaultLogLevel
		slog.Error("Validate log level:", "error", err, "default value", defaultLogLevel)
	}

	return &config{
		port:     *port,
		logLevel: *logLevel,
		secured:  *ssl,
	}
}

type config struct {
	port     int
	logLevel int
	secured  bool
}

func (c config) Port() int {
	return c.port
}

func (c config) LogLevel() int {
	return c.logLevel
}

func (c config) Secured() bool {
	return c.secured
}

func validatePort(port int) error {
	if port == 0 || port > 65535 {
		return ErrInvalidPort
	}
	return nil
}

func validateLogLevel(levelCode int) error {
	levels := map[int]slog.Level{
		-4: slog.LevelDebug,
		0:  slog.LevelInfo,
		4:  slog.LevelWarn,
		8:  slog.LevelError,
	}
	if _, ok := levels[levelCode]; !ok {
		return ErrInvalidLogLevel
	}
	return nil
}
