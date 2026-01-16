'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { createInvoice } from '@/lib/api';
import { Loader2, Receipt } from 'lucide-react';

export default function Home() {
  const router = useRouter();
  const [merchantId, setMerchantId] = useState('');
  const [amount, setAmount] = useState('');
  const [expiry, setExpiry] = useState('60');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!amount || parseFloat(amount) <= 0) {
      setError('Please enter a valid amount');
      return;
    }

    setLoading(true);
    setError('');

    try {
      const invoice = await createInvoice({
        merchant_address: merchantId || undefined,
        amount_eth: parseFloat(amount),
        expiry_minutes: parseInt(expiry) || 60,
      });
      router.push(`/invoices/${invoice.id}`);
    } catch (err: any) {
      setError(err.message || 'Something went wrong');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-gray-50 dark:bg-zinc-900 flex items-center justify-center p-4">
      <div className="w-full max-w-md bg-white dark:bg-zinc-800 rounded-2xl shadow-xl overflow-hidden border border-gray-100 dark:border-zinc-700">
        <div className="bg-blue-600 dark:bg-blue-700 p-6 text-white text-center">
          <div className="mx-auto bg-white/20 w-12 h-12 rounded-full flex items-center justify-center mb-4 backdrop-blur-sm">
            <Receipt className="w-6 h-6 text-white" />
          </div>
          <h1 className="text-2xl font-bold">Crypto Invoice</h1>
          <p className="text-blue-100 text-sm mt-1">Generate a payment link instantly</p>
        </div>

        <div className="p-8">
          <form onSubmit={handleSubmit} className="space-y-6">
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                Merchant Address (Optional)
              </label>
              <input
                type="text"
                value={merchantId}
                onChange={(e) => setMerchantId(e.target.value)}
                className="w-full px-4 py-3 rounded-lg border border-gray-300 dark:border-zinc-600 bg-white dark:bg-zinc-900 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all outline-none text-sm font-mono"
                placeholder="0x..."
              />
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  Amount (TICS)
                </label>
                <div className="relative">
                  <input
                    type="number"
                    step="0.000000000000000001"
                    value={amount}
                    onChange={(e) => setAmount(e.target.value)}
                    className="w-full px-4 py-3 pl-12 rounded-lg border border-gray-300 dark:border-zinc-600 bg-white dark:bg-zinc-900 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all outline-none font-mono"
                    placeholder="0.05"
                    required
                  />
                  <span className="absolute left-4 top-3.5 text-gray-400 font-bold text-sm">TICS</span>
                </div>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  Expires In (Mins)
                </label>
                <input
                  type="number"
                  value={expiry}
                  onChange={(e) => setExpiry(e.target.value)}
                  className="w-full px-4 py-3 rounded-lg border border-gray-300 dark:border-zinc-600 bg-white dark:bg-zinc-900 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all outline-none font-mono"
                  placeholder="60"
                  required
                />
              </div>
            </div>

            <button
              type="submit"
              disabled={loading}
              className="w-full mt-6 bg-blue-600 hover:bg-blue-700 text-white font-bold py-3 px-4 rounded-lg transform transition-all duration-200 hover:scale-[1.02] active:scale-[0.98] disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2 shadow-lg hover:shadow-blue-500/30"
            >
              {loading ? (
                <>
                  <Loader2 className="w-5 h-5 animate-spin" />
                  <span>Creating Invoice...</span>
                </>
              ) : (
                <span>Generate Invoice</span>
              )}
            </button>
          </form>
        </div>
        
        <div className="px-8 pb-6 text-center">
          <p className="text-xs text-gray-400">
            Powered by Crypto Invoice Generator
          </p>
        </div>
      </div>
    </div>
  );
}
