package config

import "os"

type PaymentConfig struct {
	Address           string
	InvoiceExpiryMins int
}

func LoadPaymentConfig() *PaymentConfig {
	return &PaymentConfig{
		Address:           os.Getenv("PAYMENT_ADDRESS"),
		InvoiceExpiryMins: 5,
	}
}
