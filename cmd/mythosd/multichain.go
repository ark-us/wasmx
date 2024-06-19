package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cobra"

	tmconfig "github.com/cometbft/cometbft/config"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/types/module"

	cosmosmodtypes "mythos/v1/x/cosmosmod/types"
	// "mythos/v1/testutil/network"
)

type internalVal struct {
	Amount              string `json:"amount"`
	Moniker             string `json:"moniker"`
	Identity            string `json:"identity,omitempty"`
	Website             string `json:"website,omitempty"`
	Security            string `json:"security,omitempty"`
	Details             string `json:"details,omitempty"`
	CommissionRate      string `json:"commission-rate"`
	CommissionMaxRate   string `json:"commission-max-rate"`
	CommissionMaxChange string `json:"commission-max-change-rate"`
	MinSelfDelegation   string `json:"min-self-delegation"`
}

func testnetCreateHierarchy(
	clientCtx client.Context,
	cmd *cobra.Command,
	nodeConfig *tmconfig.Config,
	mbm module.BasicManager,
	genBalIterator cosmosmodtypes.GenesisBalancesIterator,
	args initArgs,
	maxLevel int,
	validatorPerLevelCount int,
) error {
	level1Chains := int(math.Pow(float64(validatorPerLevelCount), float64(maxLevel-1)))
	fmt.Printf("* creating up to %d levels, with %d validators per chain: up to %d subchains\n", maxLevel, validatorPerLevelCount, level1Chains)
	subchainIds, err := getSubChainIds(clientCtx, args, 0)
	if err != nil {
		return err
	}
	startIndex := len(subchainIds)
	fmt.Printf("* creating %d subchains; total: %d subchains; already created: %d subchains\n", level1Chains-startIndex, level1Chains, startIndex)
	// create all level1 chains
	for i := startIndex; i < level1Chains; i++ {
		nodeIndex := i * validatorPerLevelCount
		subChainId, err := registerLevelChain(clientCtx, cmd, nodeConfig, mbm, genBalIterator, args, i, nodeIndex, startIndex)
		if err != nil {
			return err
		}
		time.Sleep(time.Second * 5)
		for j := 0; j < validatorPerLevelCount; j++ {
			err = registerLevelChainValidator(clientCtx, cmd, nodeConfig, mbm, genBalIterator, args, subChainId, nodeIndex+j, j)
			if err != nil {
				return err
			}
			time.Sleep(time.Second * 3)
		}
		err = initializeLevelChain(clientCtx, cmd, nodeConfig, mbm, genBalIterator, args, subChainId, i, nodeIndex)
		if err != nil {
			return err
		}
		time.Sleep(time.Second * 5)
	}
	return nil
}

func registerLevelChain(
	clientCtx client.Context,
	cmd *cobra.Command,
	nodeConfig *tmconfig.Config,
	mbm module.BasicManager,
	genBalIterator cosmosmodtypes.GenesisBalancesIterator,
	args initArgs,
	chainIndex int,
	nodeIndex int,
	chainCount int,
) (string, error) {
	chainBaseName := fmt.Sprintf("chain%d", chainIndex)
	denomUnit := fmt.Sprintf("lvl%d", chainIndex)
	nodeName := fmt.Sprintf("node%d", nodeIndex)
	nodeDir := fmt.Sprintf("%s/node%d/%s", clientCtx.HomeDir, nodeIndex, args.nodeDaemonHome)

	cmdargs := []string{
		"tx",
		"multichain",
		"register-subchain",
		chainBaseName,
		denomUnit,
		"18",
		"1", // level 1
		"10000000000000000000",
		fmt.Sprintf("--chain-id=%s", args.chainID),
		fmt.Sprintf("--from=%s", nodeName),
		fmt.Sprintf("--keyring-backend=%s", args.keyringBackend),
		fmt.Sprintf("--home=%s", nodeDir),
		fmt.Sprintf("--fees=90000000000%s", "amyt"),
		fmt.Sprintf("--gas=90000000"),
		fmt.Sprintf("--yes"),
	}
	output, errOutput, err := executeCommand(args.nodeDaemonHome, cmdargs)
	if err != nil {
		return "", err
	}
	fmt.Println(output.String())
	fmt.Println(errOutput.String())
	txhash := extractTxHashFromOutput(output.String())
	fmt.Println("* txhash: ", txhash)

	time.Sleep(time.Second * 3)
	subchainIds, err := getSubChainIds(clientCtx, args, nodeIndex)
	if err != nil {
		return "", err
	}
	for {
		if chainCount < len(subchainIds) {
			break
		}
		time.Sleep(time.Second * 2)
		subchainIds, err = getSubChainIds(clientCtx, args, nodeIndex)
		if err != nil {
			return "", err
		}
	}
	subchainId := subchainIds[len(subchainIds)-1]
	fmt.Println("* subchainId: ", subchainId)
	return subchainId, nil
}

func getSubChainIds(
	clientCtx client.Context,
	args initArgs,
	nodeIndex int,
) ([]string, error) {
	nodeName := fmt.Sprintf("node%d", nodeIndex)
	nodeDir := fmt.Sprintf("%s/node%d/%s", clientCtx.HomeDir, nodeIndex, args.nodeDaemonHome)

	cmdargs := []string{
		"query",
		"multichain",
		"subchains",
		fmt.Sprintf("--chain-id=%s", args.chainID),
		fmt.Sprintf("--from=%s", nodeName),
		fmt.Sprintf("--keyring-backend=%s", args.keyringBackend),
		fmt.Sprintf("--home=%s", nodeDir),
	}
	output, errOutput, err := executeCommand(args.nodeDaemonHome, cmdargs)
	if err != nil {
		return nil, err
	}
	fmt.Println(output.String())
	fmt.Println(errOutput.String())
	outputStr := strings.TrimSpace(output.String())
	subchainIds := []string{}
	err = json.Unmarshal([]byte(outputStr), &subchainIds)
	if err != nil {
		return nil, err
	}
	fmt.Println("* chainIds: ", subchainIds)
	return subchainIds, nil
}

func getSubChainValidatorAddresses(
	clientCtx client.Context,
	args initArgs,
	nodeIndex int,
	subchainId string,
) ([]string, error) {
	nodeName := fmt.Sprintf("node%d", nodeIndex)
	nodeDir := fmt.Sprintf("%s/node%d/%s", clientCtx.HomeDir, nodeIndex, args.nodeDaemonHome)

	cmdargs := []string{
		"query",
		"multichain",
		"validators",
		subchainId,
		fmt.Sprintf("--chain-id=%s", args.chainID),
		fmt.Sprintf("--from=%s", nodeName),
		fmt.Sprintf("--keyring-backend=%s", args.keyringBackend),
		fmt.Sprintf("--home=%s", nodeDir),
	}
	output, errOutput, err := executeCommand(args.nodeDaemonHome, cmdargs)
	if err != nil {
		return nil, err
	}
	fmt.Println(output.String())
	fmt.Println(errOutput.String())
	outputStr := strings.TrimSpace(output.String())
	addrs := []string{}
	err = json.Unmarshal([]byte(outputStr), &addrs)
	if err != nil {
		return nil, err
	}
	fmt.Println("* validators: ", addrs)
	return addrs, nil
}

func registerLevelChainValidator(
	clientCtx client.Context,
	cmd *cobra.Command,
	nodeConfig *tmconfig.Config,
	mbm module.BasicManager,
	genBalIterator cosmosmodtypes.GenesisBalancesIterator,
	args initArgs,
	subChainId string,
	nodeIndex int,
	validatorCount int,
) error {
	nodeName := fmt.Sprintf("node%d", nodeIndex)
	nodeDir := fmt.Sprintf("%s/node%d/%s", clientCtx.HomeDir, nodeIndex, args.nodeDaemonHome)
	pathToValidatorJson := fmt.Sprintf("%s/node%d/validator.json", clientCtx.HomeDir, nodeIndex)
	val := &internalVal{
		Amount:              "100000000000000000", // smaller than overall balance
		Moniker:             nodeName,
		Identity:            "identity",
		Website:             "website",
		Security:            "security",
		Details:             "details",
		CommissionRate:      "0.05",
		CommissionMaxRate:   "0.2",
		CommissionMaxChange: "0.05",
		MinSelfDelegation:   "1000000000000",
	}
	createValidatorJson(pathToValidatorJson, val)

	cmdargs := []string{
		"tx",
		"multichain",
		"register-subchain-validator",
		subChainId,
		pathToValidatorJson,
		fmt.Sprintf("--chain-id=%s", args.chainID),
		fmt.Sprintf("--from=%s", nodeName),
		fmt.Sprintf("--keyring-backend=%s", args.keyringBackend),
		fmt.Sprintf("--home=%s", nodeDir),
		fmt.Sprintf("--fees=90000000000%s", "amyt"),
		fmt.Sprintf("--gas=90000000"),
		fmt.Sprintf("--yes"),
	}
	output, errOutput, err := executeCommand(args.nodeDaemonHome, cmdargs)
	if err != nil {
		return err
	}
	fmt.Println(output.String())
	fmt.Println(errOutput.String())

	time.Sleep(time.Second * 3)
	addrs, err := getSubChainValidatorAddresses(clientCtx, args, nodeIndex, subChainId)
	if err != nil {
		return err
	}
	for {
		if validatorCount < len(addrs) {
			break
		}
		time.Sleep(time.Second * 2)
		addrs, err = getSubChainValidatorAddresses(clientCtx, args, nodeIndex, subChainId)
		if err != nil {
			return err
		}
	}
	fmt.Println("* subchain validators: ", addrs)
	return nil
}

func initializeLevelChain(
	clientCtx client.Context,
	cmd *cobra.Command,
	nodeConfig *tmconfig.Config,
	mbm module.BasicManager,
	genBalIterator cosmosmodtypes.GenesisBalancesIterator,
	args initArgs,
	subChainId string,
	chainIndex int,
	nodeIndex int,
) error {
	nodeName := fmt.Sprintf("node%d", nodeIndex)
	nodeDir := fmt.Sprintf("%s/node%d/%s", clientCtx.HomeDir, nodeIndex, args.nodeDaemonHome)
	cmdargs := []string{
		"tx",
		"multichain",
		"init-subchain",
		subChainId,
		fmt.Sprintf("--chain-id=%s", args.chainID),
		fmt.Sprintf("--from=%s", nodeName),
		fmt.Sprintf("--keyring-backend=%s", args.keyringBackend),
		fmt.Sprintf("--home=%s", nodeDir),
		fmt.Sprintf("--fees=90000000000%s", "amyt"),
		fmt.Sprintf("--gas=90000000"),
		fmt.Sprintf("--yes"),
	}
	output, errOutput, err := executeCommand(args.nodeDaemonHome, cmdargs)
	if err != nil {
		return err
	}
	fmt.Println(output.String())
	fmt.Println(errOutput.String())
	return nil
}

func createValidatorJson(pathToFile string, val *internalVal) error {
	bz, err := json.Marshal(val)
	if err != nil {
		return nil
	}
	return os.WriteFile(pathToFile, bz, 0o644)
}

func extractTxHashFromOutput(msg string) string {
	re := regexp.MustCompile(`txhash:\s+([A-F0-9]+)`)
	match := re.FindStringSubmatch(msg)
	if len(match) < 2 {
		fmt.Println("txhash not found")
		return ""
	}
	return match[1]
}

func executeCommand(cmdstr string, args []string) (bytes.Buffer, bytes.Buffer, error) {
	fmt.Println("==== execute command ====")
	fmt.Println(cmdstr, strings.Join(args, " "))
	fmt.Println("========")
	cmd := exec.Command(cmdstr, args...)
	// Create buffers to capture stdout and stderr
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	// Run the command
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Standard Output: %s\n", outBuf.String())
		fmt.Printf("Standard Error: %s\n", errBuf.String())
	}
	return outBuf, errBuf, err
}
