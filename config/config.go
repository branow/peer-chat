package config

import (
	"flag"
	"sync"
)

var (
	cfg  *config
	once sync.Once
)

func GetConfing() config {
	once.Do(func() {
		cfg = initConfig()
	})
	return *cfg
}

func initConfig() *config {
	port := flag.Uint("p", 8080, "Server port")
	log := flag.Int("log", 0, "Log Level")
	ssl := flag.Bool("s", true, "Secured connection")
	flag.Parse()

	return &config{
		port:     *port,
		logLevel: *log,
		secured:  *ssl,
	}
}

type config struct {
	port     uint
	logLevel int
	secured  bool
}

func (c config) Port() uint {
	return c.port
}

func (c config) LogLevel() int {
	return c.logLevel
}

func (c config) Secured() bool {
	return c.secured
}
