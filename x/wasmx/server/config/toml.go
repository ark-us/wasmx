package config

// DefaultConfigTemplate defines the configuration template for the JSON-RPC configuration
const DefaultConfigTemplate = `
###############################################################################
###                             JSON-RPC Configuration                           ###
###############################################################################

[json-rpc]

# Enable defines if the JSON-RPC web server should be enabled.
enable = {{ .JsonRpc.Enable }}

# Address defines the JSON-RPC server address to bind to.
address = "{{ .JsonRpc.Address }}"

# Address defines the EVM WebSocket server address to bind to.
ws-address = "{{ .JsonRpc.WsAddress }}"

# API defines a list of JSON-RPC namespaces that should be enabled
# Example: "eth,net,debug,web3"
api = "{{range $index, $elmt := .JsonRpc.API}}{{if $index}},{{$elmt}}{{else}}{{$elmt}}{{end}}{{end}}"

# EVMTimeout is the global timeout for eth_call. Default: 5s.
evm-timeout = "{{ .JsonRpc.EVMTimeout }}"

# HTTPTimeout is the read/write timeout of http json-rpc server.
http-timeout = "{{ .JsonRpc.HTTPTimeout }}"

# HTTPIdleTimeout is the idle timeout of http json-rpc server.
http-idle-timeout = "{{ .JsonRpc.HTTPIdleTimeout }}"

# AllowUnprotectedTxs restricts unprotected (non EIP155 signed) transactions to be submitted via
# the node's RPC when the global parameter is disabled.
allow-unprotected-txs = {{ .JsonRpc.AllowUnprotectedTxs }}

# MaxOpenConnections sets the maximum number of simultaneous connections
# for the server listener.
max-open-connections = {{ .JsonRpc.MaxOpenConnections }}

`
