import React, { useState, useEffect } from 'react';
import { 
  getMySubscription, 
  getSubscriptionPlans, 
  getMyUsage,
  createCryptoPayment,
  getPaymentStatus,
  completeMockPayment
} from '../services/api';
import { Subscription, SubscriptionPlan, UserUsage, CryptoCurrency } from '../types';

interface SubscriptionPanelProps {
  onClose: () => void;
}

const SubscriptionPanel: React.FC<SubscriptionPanelProps> = ({ onClose }) => {
  const [subscription, setSubscription] = useState<Subscription | null>(null);
  const [currentPlan, setCurrentPlan] = useState<SubscriptionPlan | null>(null);
  const [plans, setPlans] = useState<SubscriptionPlan[]>([]);
  const [currencies, setCurrencies] = useState<CryptoCurrency[]>([]);
  const [usage, setUsage] = useState<UserUsage | null>(null);
  const [limits, setLimits] = useState<{ processing_jobs: number; max_file_size: number } | null>(null);
  const [loading, setLoading] = useState(true);
  const [selectedPlan, setSelectedPlan] = useState<string>('');
  const [selectedCurrency, setSelectedCurrency] = useState<string>('BTC');
  const [paymentLoading, setPaymentLoading] = useState(false);
  const [paymentData, setPaymentData] = useState<any>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    try {
      setLoading(true);
      setError(null);

      const [subRes, plansRes, usageRes] = await Promise.all([
        getMySubscription(),
        getSubscriptionPlans(),
        getMyUsage(),
      ]);

      setSubscription(subRes.subscription);
      setCurrentPlan(subRes.plan);
      setPlans(plansRes.plans);
      setCurrencies(plansRes.currencies);
      setUsage(usageRes.usage);
      setLimits(usageRes.limits);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load subscription data');
    } finally {
      setLoading(false);
    }
  };

  const handleUpgrade = async () => {
    if (!selectedPlan) return;

    try {
      setPaymentLoading(true);
      setError(null);

      const payment = await createCryptoPayment(selectedPlan, selectedCurrency);
      setPaymentData(payment);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create payment');
    } finally {
      setPaymentLoading(false);
    }
  };

  const handleMockPayment = async () => {
    if (!paymentData?.payment_id) return;

    try {
      setPaymentLoading(true);
      await completeMockPayment(paymentData.payment_id);
      
      // Reload subscription data
      setTimeout(() => {
        loadData();
        setPaymentData(null);
        setPaymentLoading(false);
      }, 1000);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to complete payment');
      setPaymentLoading(false);
    }
  };

  const formatBytes = (bytes: number) => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
    });
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'active': return 'text-green-600 bg-green-50 dark:text-green-400 dark:bg-green-900/20';
      case 'expired': return 'text-red-600 bg-red-50 dark:text-red-400 dark:bg-red-900/20';
      case 'cancelled': return 'text-gray-600 bg-gray-50 dark:text-gray-400 dark:bg-gray-900/20';
      default: return 'text-yellow-600 bg-yellow-50 dark:text-yellow-400 dark:bg-yellow-900/20';
    }
  };

  if (loading) {
    return (
      <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
        <div className="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 shadow-xl max-w-2xl w-full mx-4 p-6">
          <div className="text-center">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600 mx-auto mb-4"></div>
            <p>Loading subscription data...</p>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 shadow-xl max-w-4xl w-full mx-4 max-h-[90vh] overflow-y-auto">
        {/* Header */}
        <div className="px-6 py-4 border-b border-gray-200 dark:border-gray-800 flex items-center justify-between">
          <h2 className="text-xl font-semibold text-gray-900 dark:text-gray-100">
            Subscription Management
          </h2>
          <button
            onClick={onClose}
            className="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
          >
            <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <div className="p-6">
          {error && (
            <div className="mb-6 p-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg">
              <p className="text-red-600 dark:text-red-400">{error}</p>
            </div>
          )}

          {/* Current Subscription */}
          {subscription && currentPlan && (
            <div className="mb-6">
              <h3 className="text-lg font-medium text-gray-900 dark:text-gray-100 mb-4">
                Current Subscription
              </h3>
              <div className="bg-gray-50 dark:bg-gray-800 rounded-lg p-4">
                <div className="flex items-center justify-between mb-2">
                  <span className="text-lg font-semibold">{currentPlan.name}</span>
                  <span className={`px-2 py-1 text-xs font-medium rounded-full ${getStatusColor(subscription.status)}`}>
                    {subscription.status}
                  </span>
                </div>
                <p className="text-gray-600 dark:text-gray-400 mb-2">{currentPlan.description}</p>
                <div className="text-sm text-gray-500 dark:text-gray-400">
                  <p>Active until: {formatDate(subscription.end_date)}</p>
                  {currentPlan.price_usd > 0 && (
                    <p>Price: ${currentPlan.price_usd}/month</p>
                  )}
                </div>
              </div>
            </div>
          )}

          {/* Usage Statistics */}
          {usage && limits && (
            <div className="mb-6">
              <h3 className="text-lg font-medium text-gray-900 dark:text-gray-100 mb-4">
                Current Usage
              </h3>
              <div className="bg-gray-50 dark:bg-gray-800 rounded-lg p-4">
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <div>
                    <div className="text-sm text-gray-600 dark:text-gray-400">Processing Jobs</div>
                    <div className="text-2xl font-semibold">
                      {usage.processing_jobs}
                      {limits.processing_jobs !== -1 && (
                        <span className="text-sm text-gray-500">/{limits.processing_jobs}</span>
                      )}
                    </div>
                    {limits.processing_jobs !== -1 && (
                      <div className="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2 mt-1">
                        <div 
                          className="bg-blue-600 h-2 rounded-full" 
                          style={{ width: `${Math.min(100, (usage.processing_jobs / limits.processing_jobs) * 100)}%` }}
                        ></div>
                      </div>
                    )}
                  </div>
                  <div>
                    <div className="text-sm text-gray-600 dark:text-gray-400">Max File Size</div>
                    <div className="text-2xl font-semibold">
                      {formatBytes(limits.max_file_size)}
                    </div>
                  </div>
                </div>
              </div>
            </div>
          )}

          {/* Payment Processing */}
          {paymentData && (
            <div className="mb-6">
              <h3 className="text-lg font-medium text-gray-900 dark:text-gray-100 mb-4">
                Payment Instructions
              </h3>
              <div className="bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-lg p-4">
                <div className="mb-4">
                  <p className="text-sm text-gray-600 dark:text-gray-400 mb-2">
                    Send exactly this amount to the address below:
                  </p>
                  <div className="bg-white dark:bg-gray-800 rounded p-3 border">
                    <div className="font-mono text-lg font-semibold">
                      {paymentData.crypto_amount} {paymentData.currency}
                    </div>
                  </div>
                </div>
                <div className="mb-4">
                  <p className="text-sm text-gray-600 dark:text-gray-400 mb-2">
                    Payment Address:
                  </p>
                  <div className="bg-white dark:bg-gray-800 rounded p-3 border">
                    <div className="font-mono text-sm break-all">
                      {paymentData.crypto_address}
                    </div>
                  </div>
                </div>
                {paymentData.payment_url.includes('localhost') && (
                  <div className="mt-4">
                    <button
                      onClick={handleMockPayment}
                      disabled={paymentLoading}
                      className="w-full bg-green-600 hover:bg-green-700 disabled:bg-gray-300 text-white py-2 px-4 rounded-lg transition-colors"
                    >
                      {paymentLoading ? 'Processing...' : 'Complete Mock Payment (Dev Only)'}
                    </button>
                  </div>
                )}
              </div>
            </div>
          )}

          {/* Available Plans */}
          {!paymentData && (
            <div>
              <h3 className="text-lg font-medium text-gray-900 dark:text-gray-100 mb-4">
                Available Plans
              </h3>
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
                {plans.map((plan) => (
                  <div
                    key={plan.id}
                    className={`border rounded-lg p-4 cursor-pointer transition-colors ${
                      selectedPlan === plan.id
                        ? 'border-blue-500 bg-blue-50 dark:bg-blue-900/20'
                        : 'border-gray-200 dark:border-gray-700 hover:border-gray-300 dark:hover:border-gray-600'
                    }`}
                    onClick={() => setSelectedPlan(plan.id)}
                  >
                    <div className="text-lg font-semibold mb-2">{plan.name}</div>
                    <div className="text-2xl font-bold text-blue-600 mb-2">
                      {plan.price_usd === 0 ? 'Free' : `$${plan.price_usd}`}
                      {plan.price_usd > 0 && <span className="text-sm text-gray-500">/month</span>}
                    </div>
                    <div className="text-sm text-gray-600 dark:text-gray-400 mb-3">
                      {plan.description}
                    </div>
                    <ul className="text-sm space-y-1">
                      {plan.features.map((feature, idx) => (
                        <li key={idx} className="flex items-start">
                          <span className="text-green-500 mr-2">✓</span>
                          {feature}
                        </li>
                      ))}
                    </ul>
                  </div>
                ))}
              </div>

              {selectedPlan && selectedPlan !== subscription?.plan_type && (
                <div className="mt-6 pt-6 border-t border-gray-200 dark:border-gray-700">
                  <h4 className="text-lg font-medium text-gray-900 dark:text-gray-100 mb-4">
                    Payment Options
                  </h4>
                  <div className="flex flex-wrap gap-4 mb-4">
                    {currencies.map((currency) => (
                      <label key={currency.code} className="flex items-center">
                        <input
                          type="radio"
                          name="currency"
                          value={currency.code}
                          checked={selectedCurrency === currency.code}
                          onChange={(e) => setSelectedCurrency(e.target.value)}
                          className="mr-2"
                        />
                        <span>{currency.symbol} {currency.name}</span>
                      </label>
                    ))}
                  </div>
                  <button
                    onClick={handleUpgrade}
                    disabled={paymentLoading || !selectedPlan}
                    className="bg-blue-600 hover:bg-blue-700 disabled:bg-gray-300 text-white py-2 px-6 rounded-lg transition-colors"
                  >
                    {paymentLoading ? 'Creating Payment...' : 'Upgrade Subscription'}
                  </button>
                </div>
              )}
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default SubscriptionPanel;
