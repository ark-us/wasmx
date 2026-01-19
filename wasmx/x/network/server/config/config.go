package config

import "strings"

const (
	DefaultNetworkEnable  = true
	DefaultNetworkLeader  = false
	DefaultNetworkPort    = "8090"
	DefaultNetworkAddress = "0.0.0.0:" + DefaultNetworkPort
	DefaultNetworkIps     = DefaultNetworkAddress
	DefaultNodeId         = "0"

	// DefaultMaxOpenConnections represents the amount of open connections (unlimited = 0)
	DefaultMaxOpenConnections = 0
)

var DefaultInitialChains = []string{}

// NetworkConfig defines the application configuration values for Network module.
type NetworkConfig struct {
	Enable  bool   `mapstructure:"enable"`
	Address string `mapstructure:"address"`
	// MaxOpenConnections sets the maximum number of simultaneous connections
	// for the server listener.
	MaxOpenConnections int  `mapstructure:"max-open-connections"`
	Leader             bool `mapstructure:"leader"`
	// comma separated list of values
	Ips string `mapstructure:"ips"`
	// comma separated list of values for each initialized chain
	Id            string   `mapstructure:"id"`
	InitialChains []string `mapstructure:"initial-chains"`
	StartEnv      string   `mapstructure:"start-env"`
}

// DefaultEVMConfig returns the default EVM configuration
func DefaultNetworkConfigConfig() *NetworkConfig {
	return &NetworkConfig{
		Enable:             DefaultNetworkEnable,
		Address:            DefaultNetworkAddress,
		MaxOpenConnections: DefaultMaxOpenConnections,
		Leader:             DefaultNetworkLeader,
		Ips:                DefaultNetworkIps,
		Id:                 DefaultNodeId,
		InitialChains:      DefaultInitialChains,
		StartEnv:           "",
	}
}

func (c NetworkConfig) Validate() error {
	// TODO
	return nil
}

func ParseStartEnv(raw string) map[string]string {
	env := map[string]string{}
	for _, pair := range strings.Split(raw, ";") {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		parts := strings.SplitN(pair, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if key == "" {
			continue
		}
		env[key] = value
	}
	return env
}
