'use client';

import { use, useEffect, useState, useCallback } from 'react';
import { getInvoice } from '@/lib/api';
import { Invoice } from '@/lib/types';
import { QRCodeSVG } from 'qrcode.react';
import { Copy, Check, ExternalLink, Loader2, AlertCircle, CheckCircle2, Clock } from 'lucide-react';
import Link from 'next/link';

export default function InvoicePage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = use(params);
  const [invoice, setInvoice] = useState<Invoice | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    let timeoutId: NodeJS.Timeout;
    let delay = 2000;
    const maxDelay = 10000;
    const multiplier = 1.5;
    let isMounted = true;

    const poll = async () => {
      try {
        const data = await getInvoice(id);
        
        if (isMounted) {
          setInvoice(data);
          setLoading(false);
          
          if (data.status === 'PAID' || data.status === 'EXPIRED') {
            return; // Stop polling
          }
          
          // Schedule next poll
          timeoutId = setTimeout(poll, delay);
          delay = Math.min(delay * multiplier, maxDelay);
        }
      } catch (err) {
        console.error(err);
        if (isMounted) {
          setInvoice((prev) => {
            if (!prev) setError('Failed to load invoice');
            return prev;
          });
          setLoading(false);
          // Retry even on error, but back off
          timeoutId = setTimeout(poll, delay);
          delay = Math.min(delay * multiplier, maxDelay);
        }
      }
    };

    poll();

    return () => {
      isMounted = false;
      clearTimeout(timeoutId);
    };
  }, [id]); // Only re-run if ID changes

  // const copyAddress = () => { ... } // Removed as inline logic is used

  if (loading && !invoice) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50 dark:bg-zinc-900">
        <Loader2 className="w-8 h-8 animate-spin text-blue-600" />
      </div>
    );
  }

  if (error || !invoice) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50 dark:bg-zinc-900 p-4">
        <div className="bg-white dark:bg-zinc-800 p-8 rounded-2xl shadow-xl text-center max-w-md w-full">
          <AlertCircle className="w-12 h-12 text-red-500 mx-auto mb-4" />
          <h2 className="text-xl font-bold mb-2">Error</h2>
          <p className="text-gray-600 dark:text-gray-400 mb-6">{error || 'Invoice not found'}</p>
          <Link href="/" className="inline-block bg-blue-600 px-6 py-2 rounded-lg text-white font-medium hover:bg-blue-700 transition">
            Create New Invoice
          </Link>
        </div>
      </div>
    );
  }

  const isPaid = invoice.status === 'PAID';
  const isExpired = invoice.status === 'EXPIRED';

  return (
    <div className="min-h-screen bg-gray-50 dark:bg-zinc-900 flex items-center justify-center p-4">
      <div className="w-full max-w-md bg-white dark:bg-zinc-800 rounded-2xl shadow-xl overflow-hidden border border-gray-100 dark:border-zinc-700">
        
        {/* Status Header */}
        <div className={`p-6 text-white text-center transition-colors duration-500 ${
          isPaid ? 'bg-green-600' : isExpired ? 'bg-red-500' : 'bg-blue-600'
        }`}>
          <div className="mx-auto bg-white/20 w-16 h-16 rounded-full flex items-center justify-center mb-4 backdrop-blur-sm">
            {isPaid ? (
              <CheckCircle2 className="w-8 h-8 text-white" />
            ) : isExpired ? (
              <AlertCircle className="w-8 h-8 text-white" />
            ) : (
              <Loader2 className="w-8 h-8 text-white animate-spin-slow" />
            )}
          </div>
          <h1 className="text-2xl font-bold tracking-wide">
            {isPaid ? 'PAYMENT RECEIVED' : isExpired ? 'INVOICE EXPIRED' : 'AWAITING PAYMENT'}
          </h1>
          {isPaid && (
            <a 
              href={`https://testnet.qubetics.work/tx/${invoice.tx_hash}`}
              target="_blank"
              rel="noopener noreferrer"
              className="inline-flex items-center gap-1 mt-2 text-green-100 hover:text-white underline text-sm"
            >
              View Transaction <ExternalLink className="w-3 h-3" />
            </a>
          )}
        </div>

        <div className="p-8">
          
          {/* Amount Display */}
          <div className="text-center mb-8">
            <p className="text-sm text-gray-500 dark:text-gray-400 uppercase font-semibold tracking-wider mb-1">Total Amount</p>
            <div className="flex items-end justify-center gap-2 text-gray-900 dark:text-white">
              <span className="text-4xl font-bold">{invoice.amount_eth}</span>
              <span className="text-xl font-medium mb-1.5 text-gray-500">TICS</span>
            </div>
          </div>

          {!isPaid && !isExpired && (
            <>
              {/* Payment Instructions */}
              <div className="space-y-4">
                <div className="p-4 bg-blue-50 dark:bg-blue-900/10 rounded-xl border border-blue-100 dark:border-blue-800">
                  <h3 className="text-sm font-bold text-blue-900 dark:text-blue-100 mb-2 flex items-center gap-2">
                    <AlertCircle className="w-4 h-4" />
                    How to Pay
                  </h3>
                  <p className="text-xs text-blue-700 dark:text-blue-300 leading-relaxed">
                    Call the <code className="font-mono bg-blue-100 dark:bg-blue-900/30 px-1 py-0.5 rounded">payInvoice</code> function on the smart contract with the exact ID and Amount.
                  </p>
                </div>

                {/* Contract Info */}
                <div className="space-y-3">
                  <div>
                    <label className="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                      Contract Address
                    </label>
                    <div 
                      onClick={() => {
                         navigator.clipboard.writeText(invoice.contract_address);
                         setCopied(true);
                         setTimeout(() => setCopied(false), 2000);
                      }}
                      className="group relative flex items-center justify-between p-3 bg-gray-50 dark:bg-zinc-900/50 border border-gray-200 dark:border-zinc-700 rounded-lg cursor-pointer hover:border-blue-400 dark:hover:border-blue-500 transition-colors"
                    >
                      <code className="text-xs text-gray-800 dark:text-gray-200 truncate pr-8 font-mono">
                        {invoice.contract_address}
                      </code>
                      <div className="absolute right-3 p-1.5 rounded-md text-gray-400 group-hover:text-blue-500 group-hover:bg-blue-50 dark:group-hover:bg-blue-900/20 transition-all">
                        {copied ? <Check className="w-4 h-4" /> : <Copy className="w-4 h-4" />}
                      </div>
                    </div>
                  </div>

                  <div>
                    <label className="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                      Invoice ID (uint256)
                    </label>
                    <div className="flex items-center justify-center p-4 bg-gray-50 dark:bg-zinc-900/50 border border-gray-200 dark:border-zinc-700 rounded-lg min-h-[60px]">
                      {invoice.onchain_invoice_id ? (
                        <code className="text-lg font-bold text-gray-800 dark:text-gray-200 font-mono">
                          {invoice.onchain_invoice_id}
                        </code>
                      ) : (
                        <div className="flex flex-col items-center gap-2">
                          <div className="flex items-center gap-2 text-blue-600 dark:text-blue-400">
                             <Loader2 className="w-4 h-4 animate-spin" />
                             <span className="text-sm font-medium">Syncing with blockchain...</span>
                          </div>
                          <p className="text-[10px] text-gray-400 text-center px-4 leading-tight">
                            Your transaction is being processed. The ID will appear once it's mined.
                          </p>
                        </div>
                      )}
                    </div>
                  </div>
                </div>

                <div className="pt-2">
                   <a 
                     href={`https://testnet.qubetics.work/address/${invoice.contract_address}`}
                     target="_blank"
                     rel="noopener noreferrer"
                     className="block w-full bg-blue-600 hover:bg-blue-700 text-white font-bold py-3 px-4 rounded-lg text-center transition-colors shadow-lg shadow-blue-500/20"
                   >
                     Pay via Qubetics Explorer
                   </a>
                </div>
              </div>
            </>
          )}

          {/* Details Footer */}
          <div className="mt-8 pt-6 border-t border-gray-100 dark:border-zinc-700 space-y-3">
            <div className="flex justify-between text-sm">
              <span className="text-gray-500">Invoice ID</span>
              <span className="font-mono text-gray-700 dark:text-gray-300">
                {invoice.id.slice(0, 8)}...
              </span>
            </div>
            <div className="flex justify-between text-sm">
              <span className="text-gray-500">Merchant</span>
              <span className="text-gray-700 dark:text-gray-300 font-mono text-xs">{invoice.merchant_address}</span>
            </div>
            <div className="flex justify-between text-sm">
              <span className="text-gray-500">Created</span>
              <span className="text-gray-700 dark:text-gray-300">
                {new Date(invoice.created_at).toLocaleString()}
              </span>
            </div>
            <div className="flex justify-between text-sm">
              <span className="text-gray-500">Expires</span>
              <div className="flex items-center gap-1 text-gray-700 dark:text-gray-300">
                <Clock className="w-3 h-3 text-gray-400" />
                <span>{new Date(invoice.expires_at).toLocaleTimeString()}</span>
              </div>
            </div>
          </div>

          <div className="mt-8 text-center">
            <Link href="/" className="text-sm text-blue-600 hover:text-blue-700 hover:underline">
              Create another invoice
            </Link>
          </div>

        </div>
      </div>
    </div>
  );
}
