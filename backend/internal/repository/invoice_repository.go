package repository

import (
	"time"

	"github.com/user/crypto-invoice-generator/backend/internal/models"
	"gorm.io/gorm"
)

type InvoiceRepository interface {
	Create(invoice *models.Invoice) error
	FindByID(id string) (*models.Invoice, error)
	FindByOnchainID(onchainID string) (*models.Invoice, error)
	FindByTxHash(txHash string) (*models.Invoice, error)
	UpdateStatus(id string, status models.InvoiceStatus, txHash string, payer string) error
	UpdateOnchainID(id string, onchainID string) error
	FindPending() ([]models.Invoice, error)
	UpdateExpired(now time.Time) error
}

type invoiceRepository struct {
	db *gorm.DB
}

func NewInvoiceRepository(db *gorm.DB) InvoiceRepository {
	return &invoiceRepository{db: db}
}

func (r *invoiceRepository) Create(invoice *models.Invoice) error {
	return r.db.Create(invoice).Error
}

func (r *invoiceRepository) FindByID(id string) (*models.Invoice, error) {
	var invoice models.Invoice
	if err := r.db.Where("id = ?", id).First(&invoice).Error; err != nil {
		return nil, err
	}
	return &invoice, nil
}

func (r *invoiceRepository) FindByOnchainID(onchainID string) (*models.Invoice, error) {
	var invoice models.Invoice
	if err := r.db.Where("onchain_invoice_id = ?", onchainID).First(&invoice).Error; err != nil {
		return nil, err
	}
	return &invoice, nil
}

func (r *invoiceRepository) FindByTxHash(txHash string) (*models.Invoice, error) {
	var invoice models.Invoice
	if err := r.db.Where("tx_hash = ?", txHash).First(&invoice).Error; err != nil {
		return nil, err
	}
	return &invoice, nil
}

func (r *invoiceRepository) UpdateStatus(id string, status models.InvoiceStatus, txHash string, payer string) error {
	updates := map[string]interface{}{"status": status}
	if txHash != "" {
		updates["tx_hash"] = txHash
	}
	if payer != "" {
		updates["payer_address"] = payer
	}
	return r.db.Model(&models.Invoice{}).Where("id = ?", id).Updates(updates).Error
}

func (r *invoiceRepository) UpdateOnchainID(id string, onchainID string) error {
	return r.db.Model(&models.Invoice{}).Where("id = ?", id).Update("onchain_invoice_id", onchainID).Error
}

func (r *invoiceRepository) FindPending() ([]models.Invoice, error) {
	var invoices []models.Invoice
	err := r.db.Where("status = ?", models.StatusPending).Find(&invoices).Error
	return invoices, err
}

func (r *invoiceRepository) UpdateExpired(now time.Time) error {
	return r.db.Model(&models.Invoice{}).
		Where("status = ? AND expires_at < ?", models.StatusPending, now).
		Update("status", models.StatusExpired).Error
}
