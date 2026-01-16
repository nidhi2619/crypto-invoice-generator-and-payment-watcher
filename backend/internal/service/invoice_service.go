package service

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/user/crypto-invoice-generator/backend/internal/config"
	"github.com/user/crypto-invoice-generator/backend/internal/models"
	"github.com/user/crypto-invoice-generator/backend/internal/repository"
)

// ABI file path
const abiPath = "internal/abi/invoice.json"

type InvoiceService interface {
	CreateInvoice(merchantAddr string, amountETH float64, expiryMins int) (*models.Invoice, error)
	GetInvoice(id string) (*models.Invoice, error)
}

type invoiceService struct {
	repo      repository.InvoiceRepository
	config    *config.Config
	client    *ethclient.Client
	parsedABI abi.ABI
}

func NewInvoiceService(repo repository.InvoiceRepository, cfg *config.Config, client *ethclient.Client) InvoiceService {
	abiFile, err := os.Open(abiPath)
	if err != nil {
		panic("Failed to open ABI file: " + err.Error())
	}
	defer abiFile.Close()

	parsed, err := abi.JSON(abiFile)
	if err != nil {
		panic("Failed to parse contract ABI: " + err.Error())
	}
	return &invoiceService{
		repo:      repo,
		config:    cfg,
		client:    client,
		parsedABI: parsed,
	}
}

func (s *invoiceService) CreateInvoice(merchantAddr string, amountETH float64, expiryMins int) (*models.Invoice, error) {
	// 1. Convert inputs
	// Use explicit big.Float to big.Int conversion for precision
	amountWei := new(big.Int)
	amountFloat := big.NewFloat(amountETH)
	multiplier := big.NewFloat(1e18)
	amountFloat.Mul(amountFloat, multiplier)
	amountFloat.Int(amountWei)

	expiresAt := time.Now().Add(time.Duration(expiryMins) * time.Minute)
	expiresAtUnix := big.NewInt(expiresAt.Unix())

	if merchantAddr == "" {
		// Default to owner/merchant from config if not provided
		// Assuming for MVP the backend wallet creates invoices for itself or configured merchant
		// If merchant matches config, use that.
		// NOTE: In this design, the 'owner' (backend) creates the invoice.
		// The 'merchant' param in contract is who gets paid.
		// We'll use the configured payment address if no specific merchant is requested.
		// But wait, the request asks for 'merchant_address' (optional).
		// If empty, use config.
		if s.config.Payment.Address != "" {
			merchantAddr = s.config.Payment.Address
		}
	}
	merchantCommonAddr := common.HexToAddress(merchantAddr)

	// 2. Transact with Contract
	txHash, err := s.createInvoiceOnChain(merchantCommonAddr, amountWei, expiresAtUnix)
	if err != nil {
		return nil, fmt.Errorf("failed to create invoice on-chain: %v", err)
	}

	// 3. Save to DB
	invoice := &models.Invoice{
		MerchantAddress: merchantAddr,
		AmountWei:       amountWei.String(),
		Status:          models.StatusPending,
		ExpiresAt:       expiresAt,
		ContractAddress: s.config.Ethereum.ContractAddress,
		TxHash:          &txHash,
	}

	if err := s.repo.Create(invoice); err != nil {
		return nil, err
	}

	// Populate display fields
	invoice.AmountETH = fmt.Sprintf("%f", amountETH)

	return invoice, nil
}

func (s *invoiceService) GetInvoice(id string) (*models.Invoice, error) {
	invoice, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	// Populate computed ETH amount for display
	wei, ok := new(big.Int).SetString(invoice.AmountWei, 10)
	if ok {
		ethFloat := new(big.Float).Quo(new(big.Float).SetInt(wei), big.NewFloat(1e18))
		invoice.AmountETH = fmt.Sprintf("%f", ethFloat)
	}
	invoice.ContractAddress = s.config.Ethereum.ContractAddress

	return invoice, nil
}

func (s *invoiceService) createInvoiceOnChain(merchant common.Address, amountWei *big.Int, expiresAt *big.Int) (string, error) {
	ctx := context.Background()

	// Private Key
	pkStr := strings.TrimPrefix(s.config.Ethereum.PrivateKey, "0x")
	if pkStr == "" {
		return "", fmt.Errorf("DEPLOYER_PRIVATE_KEY is missing in your configuration")
	}

	privateKey, err := crypto.HexToECDSA(pkStr)
	if err != nil {
		return "", fmt.Errorf("invalid DEPLOYER_PRIVATE_KEY: %v", err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return "", fmt.Errorf("error casting public key to ECDSA")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := s.client.PendingNonceAt(ctx, fromAddress)
	if err != nil {
		return "", fmt.Errorf("failed to get nonce: %v", err)
	}

	gasPrice, err := s.client.SuggestGasPrice(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to suggest gas price: %v", err)
	}

	chainID, err := s.client.NetworkID(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get chain ID: %v", err)
	}

	contractAddr := common.HexToAddress(s.config.Ethereum.ContractAddress)

	// Pack input data
	data, err := s.parsedABI.Pack("createInvoice", merchant, amountWei, expiresAt)
	if err != nil {
		return "", fmt.Errorf("failed to pack data: %v", err)
	}

	gasLimit := uint64(300000)
	tx := types.NewTransaction(nonce, contractAddr, big.NewInt(0), gasLimit, gasPrice, data)

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign tx: %v", err)
	}

	if err := s.client.SendTransaction(ctx, signedTx); err != nil {
		return "", fmt.Errorf("failed to send tx: %v", err)
	}

	return signedTx.Hash().Hex(), nil
}
