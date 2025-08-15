import React, { useState } from 'react';
import { useApp } from '../contexts/AppContext';
import FileSelector from './FileSelector';
import ConsoleView from './ConsoleView';
import ControlPanel from './ControlPanel';
import BatchCopyDialog from './dialogs/BatchCopyDialog';
import AddTextDialog from './dialogs/AddTextDialog';
import DeleteWatermarksDialog from './dialogs/DeleteWatermarksDialog';
import SubscriptionPanel from './SubscriptionPanel';

const HomeScreen: React.FC = () => {
  const { state } = useApp();
  const [showSubscription, setShowSubscription] = useState(false);

  return (
    <div className="flex-1 bg-gray-50 dark:bg-gray-950">
      {/* Header */}
      <div className="border-b border-gray-200 dark:border-gray-800 bg-white/80 dark:bg-gray-900/80 backdrop-blur-sm">
        <div className="max-w-7xl mx-auto px-6 py-4">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-2xl font-semibold text-gray-900 dark:text-gray-100">
                Photo Processing
              </h1>
              <p className="text-sm text-gray-600 dark:text-gray-400 mt-1">
                Encrypt, decrypt, and watermark your photos with professional tools
              </p>
            </div>
            <div className="flex items-center gap-2">
              <button
                onClick={() => setShowSubscription(true)}
                className="px-3 py-1.5 text-sm bg-blue-600 hover:bg-blue-700 text-white rounded-lg transition-colors"
              >
                Subscription
              </button>
              {state.processing.isProcessing && (
                <div className="flex items-center gap-2 px-3 py-1.5 bg-blue-50 dark:bg-blue-900/20 rounded-lg">
                  <div className="w-4 h-4 border-2 border-blue-600 border-t-transparent rounded-full animate-spin"></div>
                  <span className="text-sm text-blue-700 dark:text-blue-300">Processing...</span>
                </div>
              )}
            </div>
          </div>
        </div>
      </div>

      {/* Main Content */}
      <div className="max-w-7xl mx-auto px-6 py-6">
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          {/* Left Column - File Selection & Controls */}
          <div className="lg:col-span-1 space-y-6">
            {/* File Selection Card */}
            <div className="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 shadow-sm overflow-hidden">
              <div className="px-6 py-4 border-b border-gray-200 dark:border-gray-800">
                <h2 className="text-lg font-medium text-gray-900 dark:text-gray-100">
                  File Selection
                </h2>
                <p className="text-sm text-gray-600 dark:text-gray-400 mt-1">
                  Choose files to process
                </p>
              </div>
              <div className="p-6">
                <FileSelector />
              </div>
            </div>

            {/* Control Panel Card */}
            <div className="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 shadow-sm overflow-hidden">
              <div className="px-6 py-4 border-b border-gray-200 dark:border-gray-800">
                <h2 className="text-lg font-medium text-gray-900 dark:text-gray-100">
                  Processing Tools
                </h2>
                <p className="text-sm text-gray-600 dark:text-gray-400 mt-1">
                  Configure and run operations
                </p>
              </div>
              <div className="p-6">
                <ControlPanel />
              </div>
            </div>
          </div>

          {/* Right Column - Console Output */}
          <div className="lg:col-span-2">
            <div className="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 shadow-sm overflow-hidden h-full">
              <div className="px-6 py-4 border-b border-gray-200 dark:border-gray-800">
                <h2 className="text-lg font-medium text-gray-900 dark:text-gray-100">
                  Console Output
                </h2>
                <p className="text-sm text-gray-600 dark:text-gray-400 mt-1">
                  Real-time processing logs and status
                </p>
              </div>
              <div className="flex-1 min-h-0">
                <ConsoleView />
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Dialogs */}
      {state.dialogs.batchCopy && <BatchCopyDialog />}
      {state.dialogs.addText && <AddTextDialog />}
      {state.dialogs.deleteWatermarks && <DeleteWatermarksDialog />}
      {showSubscription && <SubscriptionPanel onClose={() => setShowSubscription(false)} />}
    </div>
  );
};

export default HomeScreen;