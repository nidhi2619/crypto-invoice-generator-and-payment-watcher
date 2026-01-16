# Crypto Invoice Generator

A production-grade Crypto Invoice Generator with a self-polling payment watcher.

## Features
- **Off-chain Invoices**: Invoices are stored in the database, not on-chain.
- **On-chain Payments**: Payments are made to a single Ethereum address.
- **Self-Polling Watcher**: Background service monitors the blockchain for payments matching invoice amounts.
- **Unique Amount Strategy**: Each invoice has a unique ETH amount (base + epsilon) to identify the payer.

## Tech Stack
- **Frontend**: Next.js, TailwindCSS
- **Backend**: Go, Gin, GORM
- **Database**: PostgreSQL
- **Blockchain**: Ethereum Sepolia

## Getting Started

### Prerequisites
- Go 1.20+
- Node.js 18+
- Docker & Docker Compose

### Setup
1. Start infrastructure (Postgres, Redis):
   ```bash
   docker-compose up -d
   ```

2. Run Backend:
   ```bash
   make run
   # OR
   cd backend && go run cmd/api/main.go
   ```
   *Note: Ensure `DATABASE_URL` and `ETHEREUM_RPC` are set correctly.*

3. Run Frontend:
   ```bash
   make run-frontend
   # OR
   cd frontend && npm run dev
   ```

## API Endpoints
- `POST /api/invoices`: Create a new invoice.
- `GET /api/invoices/:id`: Get invoice status.

## Watcher Logic
The watcher runs as a background goroutine within the API binary.
1. Polls Sepolia every ~12s.
2. Scans blocks for transactions to the configured `PAYMENT_ADDRESS`.
3. Matches transaction values to pending invoices.
4. Marks invoices as `PAID` after confirmation.
