package types

import (
	"fmt"
	"strconv"
	"strings"
)

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

func GetCurrentNodeIdFromConfig(chainId string, ids string) (int, error) {
	chainids := strings.Split(ids, ";")
	for _, chainip := range chainids {
		parts := strings.Split(chainip, ":")
		if parts[0] == chainId {
			idx, err := strconv.ParseInt(parts[1], 10, 64)
			if err != nil {
				return 0, err
			}
			return int(idx), nil
		}
	}
	return 0, fmt.Errorf("chainId %s not found in config %s", chainId, ids)
}
