package config

import "os"

type Config struct {
	DB       *DBConfig
	HTTP     *HTTPConfig
	Ethereum *EthereumConfig
	Payment  *PaymentConfig
}

func NewConfig() *Config {
	return &Config{
		DB:       LoadDBConfig(),
		HTTP:     LoadHTTPConfig(),
		Ethereum: LoadEthereumConfig(),
		Payment:  LoadPaymentConfig(),
	}
}

// Helper to read environment variables
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
