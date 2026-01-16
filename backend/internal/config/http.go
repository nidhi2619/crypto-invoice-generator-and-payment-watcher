package config

type HTTPConfig struct {
	Port string
}

func LoadHTTPConfig() *HTTPConfig {
	return &HTTPConfig{
		Port: getEnv("PORT", "8080"),
	}
}
