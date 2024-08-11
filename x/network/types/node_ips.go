package types

import "strings"

func GetPeersFromConfigIps(chainId string, ips string) []string {
	ipsMapForChain := ""
	chainips := strings.Split(ips, ";")
	for _, chainip := range chainips {
		ips := strings.Split(chainip, ":")
		if ips[0] == chainId {
			ipsMapForChain = ips[1]
			break
		}
	}
	return strings.Split(ipsMapForChain, ",")
}
