import React from 'react';
import { useApp } from '../contexts/AppContext';

const ControlPanel: React.FC = () => {
  const { 
    state, 
    encrypt, 
    decrypt, 
    updateNameToInject, 
    updateAutoClearConsole, 
    toggleDialog 
  } = useApp();

  return (
    <div className="space-y-6">
      {/* Primary Actions */}
      <div className="space-y-3">
        <div className="grid grid-cols-2 gap-3">
          <button
            onClick={decrypt}
            disabled={state.processing.isProcessing}
            className="flex items-center justify-center px-4 py-3 bg-blue-600 hover:bg-blue-700 disabled:bg-gray-400 text-white rounded-xl font-medium transition-all duration-200 disabled:cursor-not-allowed shadow-sm"
          >
            <svg className="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 15v2m0 0v2m0-2h2m-2 0H9m3-7V6m0 0V4m0 2h2m-2 0H9m12 7a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            Decrypt
          </button>
          <button
            onClick={encrypt}
            disabled={state.processing.isProcessing}
            className="flex items-center justify-center px-4 py-3 bg-green-600 hover:bg-green-700 disabled:bg-gray-400 text-white rounded-xl font-medium transition-all duration-200 disabled:cursor-not-allowed shadow-sm"
          >
            <svg className="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 15v2m0 0v2m0-2h2m-2 0H9m3-7V6m0 0V4m0 2h2m-2 0H9m12 7a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            Encrypt
          </button>
        </div>
      </div>

      {/* Name to Inject */}
      <div className="space-y-2">
        <label className="block text-sm font-medium text-gray-900 dark:text-gray-100">
          Watermark Text
        </label>
        <input
          type="text"
          value={state.processing.nameToInject}
          onChange={(e) => updateNameToInject(e.target.value)}
          disabled={state.processing.isProcessing}
          className="w-full px-4 py-3 border border-gray-300 dark:border-gray-600 rounded-xl text-sm bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 placeholder-gray-500 dark:placeholder-gray-400 focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition-all duration-200 disabled:bg-gray-100 disabled:cursor-not-allowed"
          placeholder="Enter text to embed as watermark"
        />
        <p className="text-xs text-gray-600 dark:text-gray-400">
          Text will be embedded as a hidden watermark in processed files
        </p>
      </div>

      {/* Secondary Actions */}
      <div className="space-y-3">
        <div className="grid grid-cols-2 gap-3">
          <button
            onClick={() => toggleDialog('batchCopy')}
            disabled={state.processing.isProcessing}
            className="flex items-center justify-center px-4 py-2.5 bg-gray-100 dark:bg-gray-800 hover:bg-gray-200 dark:hover:bg-gray-700 text-gray-900 dark:text-gray-100 rounded-lg font-medium transition-all duration-200 disabled:bg-gray-300 disabled:cursor-not-allowed"
          >
            <svg className="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
            </svg>
            Batch Copy
          </button>
          <button
            onClick={() => toggleDialog('addText')}
            disabled={state.processing.isProcessing}
            className="flex items-center justify-center px-4 py-2.5 bg-gray-100 dark:bg-gray-800 hover:bg-gray-200 dark:hover:bg-gray-700 text-gray-900 dark:text-gray-100 rounded-lg font-medium transition-all duration-200 disabled:bg-gray-300 disabled:cursor-not-allowed"
          >
            <svg className="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 6V4m0 2a2 2 0 100 4m0-4a2 2 0 110 4m-6 8a2 2 0 100-4m0 4a2 2 0 100 4m0-4v2m0-6V4m6 6v10m6-2a2 2 0 100-4m0 4a2 2 0 100 4m0-4v2m0-6V4" />
            </svg>
            Add Text
          </button>
        </div>

        <button
          onClick={() => toggleDialog('deleteWatermarks')}
          disabled={state.processing.isProcessing}
          className="w-full flex items-center justify-center px-4 py-2.5 bg-red-50 dark:bg-red-950/20 hover:bg-red-100 dark:hover:bg-red-950/30 text-red-700 dark:text-red-300 border border-red-200 dark:border-red-800 rounded-lg font-medium transition-all duration-200 disabled:bg-gray-300 disabled:cursor-not-allowed disabled:text-gray-500"
        >
          <svg className="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
          </svg>
          Remove Watermarks
        </button>
      </div>

      {/* Settings */}
      <div className="space-y-3">
        <div className="flex items-center justify-between p-3 bg-gray-50 dark:bg-gray-800/50 rounded-lg">
          <div className="flex-1">
            <label className="text-sm font-medium text-gray-900 dark:text-gray-100">
              Auto-clear console
            </label>
            <p className="text-xs text-gray-600 dark:text-gray-400 mt-0.5">
              Automatically clear logs before each operation
            </p>
          </div>
          <label className="relative inline-flex items-center cursor-pointer">
            <input
              type="checkbox"
              checked={state.processing.autoClearConsole}
              onChange={(e) => updateAutoClearConsole(e.target.checked)}
              disabled={state.processing.isProcessing}
              className="sr-only peer"
            />
            <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-blue-300 dark:peer-focus:ring-blue-800 rounded-full peer dark:bg-gray-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-gray-600 peer-checked:bg-blue-600"></div>
          </label>
        </div>
      </div>

      {/* Progress Indicator */}
      {state.processing.isProcessing && (
        <div className="space-y-2">
          <div className="flex items-center justify-between">
            <span className="text-sm font-medium text-gray-900 dark:text-gray-100">
              Processing...
            </span>
            <span className="text-sm text-gray-600 dark:text-gray-400">
              {Math.round(state.processing.progress * 100)}%
            </span>
          </div>
          <div className="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2 overflow-hidden">
            <div
              className="bg-gradient-to-r from-blue-500 to-blue-600 h-2 rounded-full transition-all duration-300 ease-out"
              style={{ width: `${state.processing.progress * 100}%` }}
            />
          </div>
        </div>
      )}
    </div>
  );
};

export default ControlPanel;