package config

const (
	DefaultWebservEnable = true
	DefaultWebservPort   = "9999"
	// DefaultWebservAddress defines the default address to bind the websrv web server to.
	DefaultWebservAddress = "0.0.0.0:" + DefaultWebservPort

	// DefaultMaxOpenConnections represents the amount of open connections (unlimited = 0)
	DefaultMaxOpenConnections = 0
)

var DefaultCORSAllowedOrigins = []string{"*"}
var DefaultCORSAllowedMethods = []string{"GET"}
var DefaultCORSAllowedHeaders = []string{"Origin", "Accept", "Content-Type", "X-Requested-With", "X-Server-Time"}

// WebsrvConfig defines the application configuration values for websrv module.
type WebsrvConfig struct {
	Enable             bool     `mapstructure:"enable"`
	Address            string   `mapstructure:"address"`
	CORSAllowedOrigins []string `mapstructure:"cors-allowed-origins"`
	CORSAllowedMethods []string `mapstructure:"cors-allowed-methods"`
	CORSAllowedHeaders []string `mapstructure:"cors-allowed-headers"`
	// MaxOpenConnections sets the maximum number of simultaneous connections
	// for the server listener.
	MaxOpenConnections int `mapstructure:"max-open-connections"`
}

// DefaultEVMConfig returns the default EVM configuration
func DefaultWebsrvConfigConfig() *WebsrvConfig {
	return &WebsrvConfig{
		Enable:             DefaultWebservEnable,
		Address:            DefaultWebservAddress,
		CORSAllowedOrigins: DefaultCORSAllowedOrigins,
		CORSAllowedMethods: DefaultCORSAllowedMethods,
		CORSAllowedHeaders: DefaultCORSAllowedHeaders,
		MaxOpenConnections: DefaultMaxOpenConnections,
	}
}

func (c WebsrvConfig) Validate() error {
	// TODO
	return nil
}

// IsCorsEnabled returns true if cross-origin resource sharing is enabled.
func (c *WebsrvConfig) IsCorsEnabled() bool {
	return len(c.CORSAllowedOrigins) != 0
}
