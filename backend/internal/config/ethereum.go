package config

import "os"

type EthereumConfig struct {
	RPCURL          string
	ContractAddress string
	PrivateKey      string
}

func LoadEthereumConfig() *EthereumConfig {
	return &EthereumConfig{
		RPCURL:          os.Getenv("ETHEREUM_RPC"),
		ContractAddress: os.Getenv("CONTRACT_ADDRESS"),
		PrivateKey:      os.Getenv("DEPLOYER_PRIVATE_KEY"),
	}
}
