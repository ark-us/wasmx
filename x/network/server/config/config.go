package config

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
	Id string `mapstructure:"id"`
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
	}
}

func (c NetworkConfig) Validate() error {
	// TODO
	return nil
}
