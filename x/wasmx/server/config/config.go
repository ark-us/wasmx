package config

import (
	"errors"
	"fmt"
	"time"
)

const (
	DefaultJsonRpcEnable = true
	DefaultJsonRpcPort   = "8545"
	DefaultJsonRpcWsPort = "8546"
	// DefaultJsonRpcAddress defines the default address to bind the json-rpc web server to.
	DefaultJsonRpcAddress   = "0.0.0.0:" + DefaultJsonRpcPort
	DefaultJsonRpcWsAddress = "0.0.0.0:" + DefaultJsonRpcWsPort

	// DefaultEVMTimeout is the default timeout for eth_call
	DefaultEVMTimeout = 5 * time.Second

	// DefaultHTTPTimeout is the default read/write timeout of the http json-rpc server
	DefaultHTTPTimeout = 30 * time.Second

	// DefaultHTTPIdleTimeout is the default idle timeout of the http json-rpc server
	DefaultHTTPIdleTimeout = 120 * time.Second

	// DefaultAllowUnprotectedTxs value is false
	DefaultAllowUnprotectedTxs = false

	// DefaultMaxOpenConnections represents the amount of open connections (unlimited = 0)
	DefaultMaxOpenConnections = 0
)

// JsonRpcConfig defines the application configuration values for JSON RPC module.
type JsonRpcConfig struct {
	// API defines a list of JSON-RPC namespaces that should be enabled
	API       []string `mapstructure:"api"`
	Enable    bool     `mapstructure:"enable"`
	Address   string   `mapstructure:"address"`
	WsAddress string   `mapstructure:"ws-address"`
	// EVMTimeout is the global timeout for eth-call.
	EVMTimeout time.Duration `mapstructure:"evm-timeout"`
	// HTTPTimeout is the read/write timeout of http json-rpc server.
	HTTPTimeout time.Duration `mapstructure:"http-timeout"`
	// HTTPIdleTimeout is the idle timeout of http json-rpc server.
	HTTPIdleTimeout time.Duration `mapstructure:"http-idle-timeout"`
	// AllowUnprotectedTxs restricts unprotected (non EIP155 signed) transactions to be submitted via
	// the node's RPC when global parameter is disabled.
	AllowUnprotectedTxs bool `mapstructure:"allow-unprotected-txs"`
	// MaxOpenConnections sets the maximum number of simultaneous connections
	// for the server listener.
	MaxOpenConnections int `mapstructure:"max-open-connections"`
}

// GetDefaultAPINamespaces returns the default list of JSON-RPC namespaces that should be enabled
func GetDefaultAPINamespaces() []string {
	return []string{"eth"}
}

// GetAPINamespaces returns the all the available JSON-RPC API namespaces.
func GetAPINamespaces() []string {
	return []string{"eth"}
	// return []string{"web3", "eth", "personal", "net", "txpool", "debug", "miner"}
}

// DefaultEVMConfig returns the default EVM configuration
func DefaultJsonRpcConfigConfig() *JsonRpcConfig {
	return &JsonRpcConfig{
		Enable:              DefaultJsonRpcEnable,
		API:                 GetDefaultAPINamespaces(),
		Address:             DefaultJsonRpcAddress,
		WsAddress:           DefaultJsonRpcWsAddress,
		EVMTimeout:          DefaultEVMTimeout,
		HTTPTimeout:         DefaultHTTPTimeout,
		HTTPIdleTimeout:     DefaultHTTPIdleTimeout,
		AllowUnprotectedTxs: DefaultAllowUnprotectedTxs,
		MaxOpenConnections:  DefaultMaxOpenConnections,
	}
}

// Validate returns an error if the JSON-RPC configuration fields are invalid.
func (c JsonRpcConfig) Validate() error {
	if c.Enable && len(c.API) == 0 {
		return errors.New("cannot enable JSON-RPC without defining any API namespace")
	}

	if c.EVMTimeout < 0 {
		return errors.New("JSON-RPC EVM timeout duration cannot be negative")
	}

	if c.HTTPTimeout < 0 {
		return errors.New("JSON-RPC HTTP timeout duration cannot be negative")
	}

	if c.HTTPIdleTimeout < 0 {
		return errors.New("JSON-RPC HTTP idle timeout duration cannot be negative")
	}

	// check for duplicates
	seenAPIs := make(map[string]bool)
	for _, api := range c.API {
		if seenAPIs[api] {
			return fmt.Errorf("repeated API namespace '%s'", api)
		}

		seenAPIs[api] = true
	}

	return nil
}
