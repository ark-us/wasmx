package config

// DefaultConfigTemplate defines the configuration template for the Websrv configuration
const DefaultConfigTemplate = `
###############################################################################
###                             WEBSRV Configuration                           ###
###############################################################################

[websrv]

# Enable defines if the websrv web server should be enabled.
enable = {{ .Websrv.Enable }}

# Address defines the websrv HTTP server address to bind to.
address = "{{ .Websrv.Address }}"

# A list of origins a cross-domain request can be executed from
# Default value '[]' disables cors support
# Use '["*"]' to allow any origin
cors-allowed-origins = [{{ range .Websrv.CORSAllowedOrigins }}{{ printf "%q, " . }}{{end}}]

# A list of methods the client is allowed to use with cross-domain requests
cors-allowed-methods = [{{ range .Websrv.CORSAllowedMethods }}{{ printf "%q, " . }}{{end}}]

# A list of non simple headers the client is allowed to use with cross-domain requests
cors-allowed-headers = [{{ range .Websrv.CORSAllowedHeaders }}{{ printf "%q, " . }}{{end}}]

# MaxOpenConnections sets the maximum number of simultaneous connections
# for the server listener.
max-open-connections = {{ .Websrv.MaxOpenConnections }}

`
