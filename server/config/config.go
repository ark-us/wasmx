package config

import (
	"fmt"
	"path"

	"github.com/spf13/viper"

	sdkerrors "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/server/config"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"

	networkconfig "mythos/v1/x/network/server/config"
	jsonrpcconfig "mythos/v1/x/wasmx/server/config"
	websrvconfig "mythos/v1/x/websrv/server/config"
)

const (
	// DefaultGRPCAddress is the default address the gRPC server binds to.
	DefaultGRPCAddress = "0.0.0.0:9900"

	// TODO enable
	// DefaultHTTPTimeout = 30 * time.Second
	// DefaultHTTPIdleTimeout = 120 * time.Second
)

// Config defines the server's top level configuration. It includes the default app config
// from the SDK as well as the EVM configuration to enable the JSON-RPC APIs.
type Config struct {
	config.Config
	Websrv  websrvconfig.WebsrvConfig   `mapstructure:"websrv"`
	JsonRpc jsonrpcconfig.JsonRpcConfig `mapstructure:"json-rpc"`
	TLS     TLSConfig                   `mapstructure:"tls"`
	Network networkconfig.NetworkConfig `mapstructure:"network"`
}

// TLSConfig defines the certificate and matching private key for the server.
type TLSConfig struct {
	// CertificatePath the file path for the certificate .pem file
	CertificatePath string `mapstructure:"certificate-path"`
	// KeyPath the file path for the key .pem file
	KeyPath string `mapstructure:"key-path"`
}

// AppConfig helps to override default appConfig template and configs.
// return "", nil if no custom configuration is required for the application.
func AppConfig() (string, interface{}) {
	// Optionally allow the chain developer to overwrite the SDK's default
	// server config.
	srvCfg := config.DefaultConfig()

	customAppConfig := Config{
		Config:  *srvCfg,
		Websrv:  *websrvconfig.DefaultWebsrvConfigConfig(),
		JsonRpc: *jsonrpcconfig.DefaultJsonRpcConfigConfig(),
		TLS:     *DefaultTLSConfig(),
		Network: *networkconfig.DefaultNetworkConfigConfig(),
	}

	// The SDK's default minimum gas price is set to "" (empty value) inside
	// app.toml. If left empty by validators, the node will halt on startup.
	// However, the chain developer can set a default app.toml value for their
	// validators here.
	//
	// In summary:
	// - if you leave srvCfg.MinGasPrices = "", all validators MUST tweak their
	//   own app.toml config,
	// - if you set srvCfg.MinGasPrices non-empty, validators CAN tweak their
	//   own app.toml to override, or use this default value.
	//
	// In simapp, we set the min gas prices to 0.
	srvCfg.MinGasPrices = "0amyt"

	customAppTemplate := config.DefaultConfigTemplate + DefaultConfigTemplate

	return customAppTemplate, customAppConfig
}

// DefaultConfig returns server's default configuration.
func DefaultConfig() *Config {
	return &Config{
		Config:  *config.DefaultConfig(),
		Websrv:  *websrvconfig.DefaultWebsrvConfigConfig(),
		JsonRpc: *jsonrpcconfig.DefaultJsonRpcConfigConfig(),
		TLS:     *DefaultTLSConfig(),
		Network: *networkconfig.DefaultNetworkConfigConfig(),
	}
}

// DefaultTLSConfig returns the default TLS configuration
func DefaultTLSConfig() *TLSConfig {
	return &TLSConfig{
		CertificatePath: "",
		KeyPath:         "",
	}
}

// Validate returns an error if the TLS certificate and key file extensions are invalid.
func (c TLSConfig) Validate() error {
	certExt := path.Ext(c.CertificatePath)

	if c.CertificatePath != "" && certExt != ".pem" {
		return fmt.Errorf("invalid extension %s for certificate path %s, expected '.pem'", certExt, c.CertificatePath)
	}

	keyExt := path.Ext(c.KeyPath)

	if c.KeyPath != "" && keyExt != ".pem" {
		return fmt.Errorf("invalid extension %s for key path %s, expected '.pem'", keyExt, c.KeyPath)
	}

	return nil
}

// GetConfig returns a fully parsed Config object.
func GetConfig(v *viper.Viper) (Config, error) {
	cfg, err := config.GetConfig(v)
	if err != nil {
		return Config{}, err
	}
	websrvConf := websrvconfig.WebsrvConfig{
		Enable:             v.GetBool("websrv.enable"),
		EnableOAuth:        v.GetBool("websrv.enable-oauth"),
		Address:            v.GetString("websrv.address"),
		CORSAllowedOrigins: v.GetStringSlice("websrv.cors-allowed-origins"),
		CORSAllowedMethods: v.GetStringSlice("websrv.cors-allowed-methods"),
		CORSAllowedHeaders: v.GetStringSlice("websrv.cors-allowed-headers"),
		MaxOpenConnections: v.GetInt("websrv.max-open-connections"),
	}
	jsonRpcConf := jsonrpcconfig.JsonRpcConfig{
		Enable:             v.GetBool("json-rpc.enable"),
		API:                v.GetStringSlice("json-rpc.api"),
		Address:            v.GetString("json-rpc.address"),
		WsAddress:          v.GetString("json-rpc.ws-address"),
		EVMTimeout:         v.GetDuration("json-rpc.evm-timeout"),
		HTTPTimeout:        v.GetDuration("json-rpc.http-timeout"),
		HTTPIdleTimeout:    v.GetDuration("json-rpc.http-idle-timeout"),
		MaxOpenConnections: v.GetInt("json-rpc.max-open-connections"),
	}
	networkConf := networkconfig.NetworkConfig{
		Enable:             v.GetBool("network.enable"),
		Address:            v.GetString("network.address"),
		MaxOpenConnections: v.GetInt("network.max-open-connections"),
	}

	return Config{
		Config:  cfg,
		Websrv:  websrvConf,
		JsonRpc: jsonRpcConf,
		Network: networkConf,
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
		return sdkerrors.Wrapf(errortypes.ErrAppConfig, "invalid websrv config value: %s", err.Error())
	}

	if err := c.JsonRpc.Validate(); err != nil {
		return sdkerrors.Wrapf(errortypes.ErrAppConfig, "invalid json-rpc config value: %s", err.Error())
	}

	if err := c.Network.Validate(); err != nil {
		return sdkerrors.Wrapf(errortypes.ErrAppConfig, "invalid network config value: %s", err.Error())
	}

	return c.Config.ValidateBasic()
}
