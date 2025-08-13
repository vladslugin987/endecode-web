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
    <div className="bg-white dark:bg-gray-900 rounded-xl shadow-sm ring-1 ring-gray-200 dark:ring-gray-800 p-3 space-y-3">
      {/* Main Action Buttons */}
      <div className="grid grid-cols-2 gap-2">
        <button
          onClick={decrypt}
          disabled={state.processing.isProcessing}
          className="h-8 px-3 bg-blue-600 text-white rounded font-medium hover:bg-blue-700 disabled:bg-gray-400 disabled:cursor-not-allowed text-sm"
        >
          DECRYPT
        </button>
        <button
          onClick={encrypt}
          disabled={state.processing.isProcessing}
          className="h-8 px-3 bg-blue-600 text-white rounded font-medium hover:bg-blue-700 disabled:bg-gray-400 disabled:cursor-not-allowed text-sm"
        >
          ENCRYPT
        </button>
      </div>

      {/* Name to Inject Input */}
      <div>
        <label className="block text-sm text-gray-700 dark:text-gray-300 mb-1">Name to inject</label>
        <input
          type="text"
          value={state.processing.nameToInject}
          onChange={(e) => updateNameToInject(e.target.value)}
          disabled={state.processing.isProcessing}
          className="w-full h-8 px-2 border border-gray-300 dark:border-gray-700 rounded text-sm
                   bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100
                   focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent
                   disabled:bg-gray-100 disabled:cursor-not-allowed"
          placeholder="Enter name to inject"
        />
        <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
          Only latin characters, numbers and special characters
        </p>
      </div>

      {/* Additional Action Buttons */}
      <div className="grid grid-cols-2 gap-2">
        <button
          onClick={() => toggleDialog('batchCopy')}
          disabled={state.processing.isProcessing}
          className="h-8 px-3 bg-gray-100 dark:bg-gray-800 text-gray-900 dark:text-gray-100 rounded hover:bg-gray-200 dark:hover:bg-gray-700 disabled:bg-gray-300 disabled:cursor-not-allowed text-sm"
        >
          Batch Copy
        </button>
        <button
          onClick={() => toggleDialog('addText')}
          disabled={state.processing.isProcessing}
          className="h-8 px-3 bg-gray-100 dark:bg-gray-800 text-gray-900 dark:text-gray-100 rounded hover:bg-gray-200 dark:hover:bg-gray-700 disabled:bg-gray-300 disabled:cursor-not-allowed text-sm"
        >
          Add Text
        </button>
      </div>

      {/* Delete Watermarks Button */}
      <button
        onClick={() => toggleDialog('deleteWatermarks')}
        disabled={state.processing.isProcessing}
        className="w-full h-8 px-3 bg-gray-100 dark:bg-gray-800 text-gray-900 dark:text-gray-100 rounded hover:bg-gray-200 dark:hover:bg-gray-700 disabled:bg-gray-300 disabled:cursor-not-allowed text-sm"
      >
        Delete Watermarks
      </button>

      {/* Auto-clear Console Checkbox */}
      <label className="flex items-center text-sm text-gray-700 dark:text-gray-300">
        <input
          type="checkbox"
          checked={state.processing.autoClearConsole}
          onChange={(e) => updateAutoClearConsole(e.target.checked)}
          disabled={state.processing.isProcessing}
          className="mr-2 h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 dark:border-gray-600 rounded disabled:cursor-not-allowed"
        />
        Auto-clear console
      </label>

      {/* Progress Indicator */}
      {state.processing.isProcessing && (
        <div className="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2">
          <div
            className="bg-blue-600 h-2 rounded-full transition-all duration-300"
            style={{ width: `${state.processing.progress * 100}%` }}
          />
        </div>
      )}
    </div>
  );
};

export default ControlPanel;