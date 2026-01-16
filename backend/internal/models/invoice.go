package models

import (
	"time"

	"github.com/google/uuid"
)

type InvoiceStatus string

const (
	StatusPending InvoiceStatus = "PENDING"
	StatusPaid    InvoiceStatus = "PAID"
	StatusExpired InvoiceStatus = "EXPIRED"
)

type Invoice struct {
	ID               uuid.UUID     `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	OnchainInvoiceID string        `gorm:"index" json:"onchain_invoice_id,omitempty"` // uint256 as string, populated later by watcher
	MerchantAddress  string        `gorm:"not null" json:"merchant_address"`
	AmountWei        string        `gorm:"not null" json:"amount_wei"` // big.Int as string
	AmountETH        string        `gorm:"-" json:"amount_eth"`        // Computed field for display
	Status           InvoiceStatus `gorm:"type:varchar(20);default:'PENDING'" json:"status"`
	ExpiresAt        time.Time     `gorm:"not null" json:"expires_at"`
	PaymentAddress   string        `gorm:"-" json:"-"` // Deprecated, kept ignored or removed
	ContractAddress  string        `gorm:"-" json:"contract_address"`
	TxHash           *string       `gorm:"type:varchar(66)" json:"tx_hash,omitempty"`
	PayerAddress     *string       `gorm:"type:varchar(42)" json:"payer_address,omitempty"`
	CreatedAt        time.Time     `json:"created_at"`
	UpdatedAt        time.Time     `json:"updated_at"`
}

type AppState struct {
	ID                 uint   `gorm:"primaryKey" json:"id"`
	LastProcessedBlock uint64 `json:"last_processed_block"`
}
