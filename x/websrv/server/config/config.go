package config

import (
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/server/config"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	DefaultWebservEnable = true
	// DefaultWebservAddress defines the default address to bind the websrv web server to.
	DefaultWebservAddress = "0.0.0.0:80"
)

var DefaultCORSAllowedOrigins = []string{}
var DefaultCORSAllowedMethods = []string{"GET"}
var DefaultCORSAllowedHeaders = []string{"Origin", "Accept", "Content-Type", "X-Requested-With", "X-Server-Time"}

// Config defines the server's top level configuration. It includes the default app config
// from the SDK as well as the EVM configuration to enable the JSON-RPC APIs.
type Config struct {
	config.Config
	Websrv WebsrvConfig `mapstructure:"websrv"`
}

// WebsrvConfig defines the application configuration values for websrv module.
type WebsrvConfig struct {
	Enable             bool     `mapstructure:"enable"`
	Address            string   `mapstructure:"address"`
	CORSAllowedOrigins []string `mapstructure:"cors-allowed-origins"`
	CORSAllowedMethods []string `mapstructure:"cors-allowed-methods"`
	CORSAllowedHeaders []string `mapstructure:"cors-allowed-headers"`
}

// AppConfig helps to override default appConfig template and configs.
// return "", nil if no custom configuration is required for the application.
func AppConfig() (string, interface{}) {
	// Optionally allow the chain developer to overwrite the SDK's default
	// server config.
	srvCfg := config.DefaultConfig()

	customAppConfig := Config{
		Config: *srvCfg,
		Websrv: *DefaultWebsrvConfigConfig(),
	}

	customAppTemplate := config.DefaultConfigTemplate + DefaultConfigTemplate

	return customAppTemplate, customAppConfig
}

// DefaultConfig returns server's default configuration.
func DefaultConfig() *Config {
	return &Config{
		Config: *config.DefaultConfig(),
		Websrv: *DefaultWebsrvConfigConfig(),
	}
}

// DefaultEVMConfig returns the default EVM configuration
func DefaultWebsrvConfigConfig() *WebsrvConfig {
	return &WebsrvConfig{
		Enable:             DefaultWebservEnable,
		Address:            DefaultWebservAddress,
		CORSAllowedOrigins: DefaultCORSAllowedOrigins,
		CORSAllowedMethods: DefaultCORSAllowedMethods,
		CORSAllowedHeaders: DefaultCORSAllowedHeaders,
	}
}

// Validate returns an error if the tracer type is invalid.
func (c WebsrvConfig) Validate() error {
	// TODO
	return nil
}

// IsCorsEnabled returns true if cross-origin resource sharing is enabled.
func (c *WebsrvConfig) IsCorsEnabled() bool {
	return len(c.CORSAllowedOrigins) != 0
}

// GetConfig returns a fully parsed Config object.
func GetConfig(v *viper.Viper) (Config, error) {
	cfg, err := config.GetConfig(v)
	if err != nil {
		return Config{}, err
	}

	return Config{
		Config: cfg,
		Websrv: WebsrvConfig{
			Enable:             v.GetBool("websrv.enable"),
			Address:            v.GetString("websrv.address"),
			CORSAllowedOrigins: v.GetStringSlice("websrv.cors-allowed-origins"),
			CORSAllowedMethods: v.GetStringSlice("websrv.cors-allowed-methods"),
			CORSAllowedHeaders: v.GetStringSlice("websrv.cors-allowed-headers"),
		},
	}, nil
}

// ParseConfig retrieves the default environment configuration for the
// application.
func ParseConfig(v *viper.Viper) (*Config, error) {
	conf := DefaultConfig()
	err := v.Unmarshal(conf)

	return conf, err
}

// ValidateBasic returns an error any of the application configuration fields are invalid
func (c Config) ValidateBasic() error {
	if err := c.Websrv.Validate(); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrAppConfig, "invalid evm config value: %s", err.Error())
	}

	return c.Config.ValidateBasic()
}
