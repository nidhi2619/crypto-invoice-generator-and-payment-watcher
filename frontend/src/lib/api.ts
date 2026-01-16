import { Invoice, CreateInvoiceRequest } from './types';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api';

export async function createInvoice(data: CreateInvoiceRequest): Promise<Invoice> {
  const res = await fetch(`${API_BASE_URL}/invoices`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  });
  
  if (!res.ok) {
    const error = await res.json().catch(() => ({ error: 'Unknown error' }));
    throw new Error(error.error || 'Failed to create invoice');
  }
  
  return res.json();
}

export async function getInvoice(id: string): Promise<Invoice> {
  const res = await fetch(`${API_BASE_URL}/invoices/${id}`);
  
  if (!res.ok) {
    throw new Error('Failed to fetch invoice');
  }
  
  return res.json();
}
