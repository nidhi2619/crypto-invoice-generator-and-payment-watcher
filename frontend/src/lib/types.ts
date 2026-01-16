export interface Invoice {
  id: string;
  onchain_invoice_id: string;
  merchant_address: string;
  amount_wei: string;
  amount_eth: string; // Display
  contract_address: string;
  status: 'PENDING' | 'PAID' | 'EXPIRED';
  expires_at: string;
  payer_address?: string;
  tx_hash?: string;
  created_at: string;
  updated_at: string;
}

export interface CreateInvoiceRequest {
  merchant_address?: string;
  amount_eth: number;
  expiry_minutes: number;
}
