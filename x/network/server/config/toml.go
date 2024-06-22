package config

// DefaultConfigTemplate defines the configuration template for the Network configuration
const DefaultConfigTemplate = `
###############################################################################
###                             Network Configuration                           ###
###############################################################################

[network]

# Enable defines if the network GRPC server should be enabled.
enable = {{ .Network.Enable }}

# Address defines the network GRPC server address to bind to.
address = "{{ .Network.Address }}"

# MaxOpenConnections sets the maximum number of simultaneous connections
# for the server listener.
max-open-connections = {{ .Network.MaxOpenConnections }}

leader = {{ .Network.Leader }}

# Comma separated list of node ips
ips = "{{ .Network.Ips }}"

id = "{{ .Network.Id }}"

`
