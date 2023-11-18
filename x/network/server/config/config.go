package config

const (
	DefaultNetworkEnable  = true
	DefaultNetworkPort    = "8090"
	DefaultNetworkAddress = "0.0.0.0:" + DefaultNetworkPort

	// DefaultMaxOpenConnections represents the amount of open connections (unlimited = 0)
	DefaultMaxOpenConnections = 0
)

// NetworkConfig defines the application configuration values for Network module.
type NetworkConfig struct {
	Enable  bool   `mapstructure:"enable"`
	Address string `mapstructure:"address"`
	// MaxOpenConnections sets the maximum number of simultaneous connections
	// for the server listener.
	MaxOpenConnections int `mapstructure:"max-open-connections"`
}

// DefaultEVMConfig returns the default EVM configuration
func DefaultNetworkConfigConfig() *NetworkConfig {
	return &NetworkConfig{
		Enable:             DefaultNetworkEnable,
		Address:            DefaultNetworkAddress,
		MaxOpenConnections: DefaultMaxOpenConnections,
	}
}

func (c NetworkConfig) Validate() error {
	// TODO
	return nil
}
