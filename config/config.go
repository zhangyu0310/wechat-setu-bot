package config

import (
	"sync/atomic"
)

// Config Configurations of server.
type Config struct {
	SetuApiUrl     string
	WeChatUrl      string
	Intervals      uint
	R18            bool
	AtAll          bool
	PicMsg         bool
	NewsMsg        bool
	PicDownloadDir string
	PicDump        bool
	DumpServer     string
	DumpUrl        string
	SetuTransmit   bool
	TransmitServer string
	Tags           []string
	PicSize        []string
	Once           bool
	Keep           bool
}

var (
	globalConf atomic.Value
)

// InitializeConfig initialize the global config handler.
func InitializeConfig(enforceCmdArgs func(*Config)) {
	cfg := Config{}
	// Use command config cover config file.
	enforceCmdArgs(&cfg)
	StoreGlobalConfig(&cfg)
}

// GetGlobalConfig returns the global configuration for this server.
// It should store configuration from command line and configuration file.
// Other parts of the system can read the global configuration use this function.
func GetGlobalConfig() *Config {
	return globalConf.Load().(*Config)
}

// StoreGlobalConfig stores a new config to the globalConf. It mostly uses in the test to avoid some data races.
func StoreGlobalConfig(config *Config) {
	globalConf.Store(config)
}
