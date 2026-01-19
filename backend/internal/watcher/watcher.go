package watcher

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/user/crypto-invoice-generator/backend/internal/config"
	"github.com/user/crypto-invoice-generator/backend/internal/models"
	"github.com/user/crypto-invoice-generator/backend/internal/repository"
	"gorm.io/gorm"
)

const abiPath = "internal/abi/invoice.json"

type Watcher struct {
	client          *ethclient.Client
	repo            repository.InvoiceRepository
	cfg             *config.Config
	db              *gorm.DB
	contractABI     abi.ABI
	contractAddress string
}

func NewWatcher(db *gorm.DB, repo repository.InvoiceRepository, cfg *config.Config, client *ethclient.Client) *Watcher {
	abiFile, err := os.Open(abiPath)
	if err != nil {
		panic("Failed to open ABI file: " + err.Error())
	}
	defer abiFile.Close()

	parsed, err := abi.JSON(abiFile)
	if err != nil {
		panic("Failed to parse contract ABI: " + err.Error())
	}

	return &Watcher{
		client:          client,
		repo:            repo,
		cfg:             cfg,
		db:              db,
		contractABI:     parsed,
		contractAddress: cfg.Ethereum.ContractAddress,
	}
}

func (w *Watcher) Start() {
	w.startExpiryChecker()
	go func() {
		for {
			w.pollLogs()
			time.Sleep(10 * time.Second) // Poll frequently for events
		}
	}()
}

func (w *Watcher) startExpiryChecker() {
	ticker := time.NewTicker(1 * time.Minute)
	go func() {
		for range ticker.C {
			if err := w.repo.UpdateExpired(time.Now()); err != nil {
				log.Printf("Failed to update expired invoices: %v", err)
			}
		}
	}()
}

func (w *Watcher) pollLogs() {
	ctx := context.Background()

	// Get latest block
	latestBlock, err := w.client.BlockNumber(ctx)
	if err != nil {
		log.Printf("Failed to get latest block: %v", err)
		return
	}

	// Safety margin
	safeBlock := latestBlock - 2

	lastProcessed := w.getLastProcessedBlock()
	if lastProcessed >= safeBlock {
		return
	}

	startBlock := lastProcessed + 1
	// Limit range for getLogs to avoid errors (e.g. max 1000 blocks)
	if safeBlock-startBlock > 1000 {
		safeBlock = startBlock + 1000
	}

	log.Printf("Scanning logs from %d to %d", startBlock, safeBlock)

	// Filter for both InvoiceCreated and InvoicePaid
	paidID := w.contractABI.Events["InvoicePaid"].ID
	createdID := w.contractABI.Events["InvoiceCreated"].ID

	contractAddr := common.HexToAddress(w.contractAddress)

	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(int64(startBlock)),
		ToBlock:   big.NewInt(int64(safeBlock)),
		Addresses: []common.Address{contractAddr},
		Topics:    [][]common.Hash{{paidID, createdID}},
	}

	logs, err := w.client.FilterLogs(ctx, query)
	if err != nil {
		log.Printf("Failed to fetch logs: %v", err)
		return
	}

	for _, vLog := range logs {
		// Pass to the new parser method
		if err := w.parseContractEvents(ctx, nil, &types.Receipt{Logs: []*types.Log{&vLog}}, 0); err != nil {
			log.Printf("Error parsing event: %v", err)
		}
	}

	w.updateLastProcessedBlock(safeBlock)
}

func (w *Watcher) parseContractEvents(ctx context.Context, tx *types.Transaction, receipt *types.Receipt, timestamp uint64) error {
	for _, lg := range receipt.Logs {
		if len(lg.Topics) == 0 || !strings.EqualFold(lg.Address.Hex(), w.contractAddress) {
			continue
		}

		event, err := w.contractABI.EventByID(lg.Topics[0])
		if err != nil {
			continue
		}
		fmt.Println("Processing event:", event.Name)

		switch event.Name {
		case "InvoiceCreated":
			w.handleInvoiceCreated(*lg)
		case "InvoicePaid":
			w.handleInvoicePaid(*lg)
		}
	}
	return nil
}

type InvoiceCreatedEvent struct {
	AmountWei *big.Int
	ExpiresAt *big.Int
}

type InvoicePaidEvent struct {
	AmountWei *big.Int
}

func (w *Watcher) handleInvoiceCreated(vLog types.Log) {
	var raw InvoiceCreatedEvent
	if err := w.contractABI.UnpackIntoInterface(&raw, "InvoiceCreated", vLog.Data); err != nil {
		log.Printf("Failed to decode InvoiceCreated event data: %v", err)
		return
	}

	if len(vLog.Topics) < 2 {
		log.Printf("WARN: InvoiceCreated event missing indexed invoiceId")
		return
	}

	invoiceId := new(big.Int).SetBytes(vLog.Topics[1].Bytes())
	txHash := vLog.TxHash.Hex()

	log.Printf("Detected InvoiceCreated event: ID %s, Tx %s, Amount %s", invoiceId, txHash, raw.AmountWei)

	// Find invoice by TxHash
	invoice, err := w.repo.FindByTxHash(txHash)
	if err != nil {
		log.Printf("WARN: InvoiceCreated event for unknown TxHash %s", txHash)
		return
	}

	if invoice.OnchainInvoiceID != "" {
		log.Printf("Invoice %s already has OnchainID %s", invoice.ID, invoice.OnchainInvoiceID)
		return
	}

	err = w.repo.UpdateOnchainID(invoice.ID.String(), invoiceId.String())
	if err != nil {
		log.Printf("Failed to update on-chain ID for invoice %s: %v", invoice.ID, err)
	} else {
		log.Printf("Invoice %s linked to OnchainID %s", invoice.ID, invoiceId)
	}
}

func (w *Watcher) handleInvoicePaid(vLog types.Log) {
	var raw InvoicePaidEvent
	if err := w.contractABI.UnpackIntoInterface(&raw, "InvoicePaid", vLog.Data); err != nil {
		log.Printf("Failed to decode InvoicePaid event data: %v", err)
		return
	}

	if len(vLog.Topics) < 3 {
		log.Printf("WARN: InvoicePaid event missing indexed fields")
		return
	}

	invoiceId := new(big.Int).SetBytes(vLog.Topics[1].Bytes())
	payer := common.BytesToAddress(vLog.Topics[2].Bytes())

	log.Printf("Detected InvoicePaid event: ID %s from %s in tx %s, Amount %s", invoiceId, payer.Hex(), vLog.TxHash.Hex(), raw.AmountWei)

	// Find invoice in DB by on-chain ID
	invoice, err := w.repo.FindByOnchainID(invoiceId.String())
	if err != nil {
		log.Printf("WARN: InvoicePaid event for unknown on-chain ID %s", invoiceId)
		return
	}

	if invoice.Status == models.StatusPaid {
		log.Printf("Invoice %s already marked PAID", invoice.ID)
		return
	}

	err = w.repo.UpdateStatus(invoice.ID.String(), models.StatusPaid, vLog.TxHash.Hex(), payer.Hex())
	if err != nil {
		log.Printf("Failed to update invoice status: %v", err)
	} else {
		log.Printf("Invoice %s marked as PAID", invoice.ID)
	}
}

func (w *Watcher) getLastProcessedBlock() uint64 {
	var appState models.AppState
	if err := w.db.First(&appState).Error; err != nil {
		currentBlock, _ := w.client.BlockNumber(context.Background())
		// Start from now if fresh
		if currentBlock > 0 {
			currentBlock = currentBlock - 1
		}
		w.db.Create(&models.AppState{LastProcessedBlock: currentBlock})
		return currentBlock
	}
	return appState.LastProcessedBlock
}

func (w *Watcher) updateLastProcessedBlock(blockNum uint64) {
	w.db.Model(&models.AppState{}).Where("id = 1").Update("last_processed_block", blockNum)
}
