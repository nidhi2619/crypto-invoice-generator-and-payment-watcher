package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/user/crypto-invoice-generator/backend/internal/service"
)

type InvoiceHandler struct {
	service service.InvoiceService
}

func NewInvoiceHandler(service service.InvoiceService) *InvoiceHandler {
	return &InvoiceHandler{service: service}
}

type CreateInvoiceRequest struct {
	MerchantAddress string  `json:"merchant_address"` // Optional, defaults to config
	AmountETH       float64 `json:"amount_eth" binding:"required,gt=0"`
	ExpiryMinutes   int     `json:"expiry_minutes" binding:"required,gt=0"`
}

func (h *InvoiceHandler) CreateInvoice(c *gin.Context) {
	var req CreateInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	invoice, err := h.service.CreateInvoice(req.MerchantAddress, req.AmountETH, req.ExpiryMinutes)
	if err != nil {
		fmt.Printf("FAILURE: CreateInvoice failed: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, invoice)
}

func (h *InvoiceHandler) GetInvoice(c *gin.Context) {
	id := c.Param("id")
	invoice, err := h.service.GetInvoice(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Invoice not found"})
		return
	}

	c.JSON(http.StatusOK, invoice)
}
